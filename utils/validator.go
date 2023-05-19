package utils

import (
	"fmt"
	"image"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gobeam/stringy"
	"github.com/gofiber/fiber/v2"
	"github.com/igeargeek/igg-golang-api-response/response"
	"github.com/samber/lo"
)

type ErrorResponse struct {
	FailedField string `json:"failed_field"`
	Tag         string `json:"tag"`
	Value       string `json:"value"`
	Message     string `json:"message"`
}

type ImageResolution struct {
	Width  int
	Height int
}

type ValidateFile struct {
	Field           string
	IsRequired      bool
	Extension       []string
	Size            float64
	ImageResolution *ImageResolution
}

func ValidateStruct[T any](c *fiber.Ctx, files *[]ValidateFile) error {
	body := new(T)
	if err := c.BodyParser(body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Body parser error: "+err.Error())
	}

	var errors []*ErrorResponse
	validate := validator.New()
	if err := validate.Struct(body); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			errors = append(errors, &ErrorResponse{
				FailedField: stringy.New(e.Field()).SnakeCase("?", "").ToLower(),
				Tag:         e.Tag(),
				Value:       e.Param(),
			})
		}
	}

	rule := NewRule()
	if files != nil && len(*files) != 0 {
		for _, file := range *files {
			fileHeader, err := c.FormFile(file.Field)
			if file.IsRequired {
				if fileHeader == nil {
					errors = append(errors, &ErrorResponse{
						FailedField: file.Field,
						Tag:         "required",
					})
				}
			}
			if fileHeader != nil {
				if err != nil {
					return fiber.NewError(fiber.StatusInternalServerError, "File "+file.Field+" error: "+err.Error())
				}

				if errorFile := rule.File(file.Field, fileHeader, file.Extension, file.Size); errorFile != nil {
					errors = append(errors, &ErrorResponse{
						FailedField: file.Field,
						Tag:         errorFile.Tag,
						Value:       errorFile.Value,
					})
				} else {
					if file.ImageResolution != nil {
						if errorResolution := rule.Resolution(file.Field, fileHeader, file.ImageResolution.Width, file.ImageResolution.Width); errorResolution != nil {
							errors = append(errors, &ErrorResponse{
								FailedField: file.Field,
								Tag:         errorResolution.Tag,
								Value:       errorResolution.Value,
							})
						}
					}
				}
			}
		}
	}

	if len(errors) > 0 {
		status, resData := response.ValidateFailed(errors, "")
		return c.Status(status).JSON(resData)
	}

	c.Locals("request", body)
	return c.Next()
}

type rule interface {
	File(fieldName string, fileHeader *multipart.FileHeader, allowExtension []string, maxSize float64) (errorResponse *ErrorResponse)
	Resolution(fieldName string, fileHeader *multipart.FileHeader, imageWidth int, imageHeight int) (errorResponse *ErrorResponse)
}

type Rule struct{}

func NewRule() rule {
	return &Rule{}
}

func (r *Rule) File(fieldName string, fileHeader *multipart.FileHeader, allowExtension []string, maxSize float64) (errorResponse *ErrorResponse) {
	var errorElements ErrorResponse
	errorElements.FailedField = fieldName

	fileSize := fileHeader.Size
	fileName := fileHeader.Filename
	fileExtension := filepath.Ext(fileName)

	if !lo.Contains[string](allowExtension, fileExtension) {
		errorElements.Tag = "file_extension"
		errorElements.Value = strings.Join(allowExtension[:], ", ")

		return &errorElements
	}

	maxSize = maxSize * (1024 * 1024)
	if float64(fileSize) > maxSize {
		errorElements.Tag = "file_size"
		errorElements.Value = fmt.Sprintf("%f", maxSize)

		return &errorElements
	}

	return nil
}

func (r *Rule) Resolution(fieldName string, fileHeader *multipart.FileHeader, imageWidth int, imageHeight int) (errorResponse *ErrorResponse) {
	var errorElements ErrorResponse
	errorElements.FailedField = fieldName

	file, _ := fileHeader.Open()
	defer file.Close()

	image, _, err := image.DecodeConfig(file)
	if err != nil {
		errorElements.Tag = "resolution"
		errorElements.Value = fmt.Sprintf("%dx%d", imageWidth, imageHeight)

		return &errorElements
	}

	if image.Width != imageWidth || image.Height != imageHeight {
		errorElements.Tag = "resolution"
		errorElements.Value = fmt.Sprintf("%dx%d", imageWidth, imageHeight)

		return &errorElements
	}

	return nil
}
