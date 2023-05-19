# IGG Golang Standard API Utils

## Install
```command
go get github.com/igeargeek/igg-golang-api-utils
```

## How to use

```golang
package controllers

import "github.com/igeargeek/igg-golang-api-utils/utils"

## Pagination
type ExampleController struct {
	Pagination utils.IPagination
}

func NewExampleController() exampleController {
	return &ExampleController{
		Pagination: utils.NewPagination(),
	}
}

func (root *ExampleController) GetList(c *fiber.Ctx) error  {
	allowOrderField := []string{"id", "created_at"}
	pagination, err := root.Pagination.GetPagination(c, allowOrderField)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError)
	}
  example := []models.Example{}
  var total int64
  query := root.DB.Model(&models.Example{}).Count(&total).
		Limit(pagination.PerPage).
		Offset(pagination.Offset).
		Order(pagination.OrderField + " " + pagination.OrderDirection).
		Find(&example)
  if query.Error != nil {
		return fiber.NewError(fiber.StatusInternalServerError)
	}
	respData := response.Pagination{
		Data:    users,
		Total:   total,
		PerPage: int64(pagination.PerPage),
		Page:    int64(pagination.Page),
	}
	return root.Response.Paginate(c, fiber.StatusOK, "", respData)
}

```

```golang

## Middleware Validators

package validators

import (
	"github.com/gofiber/fiber/v2"
  "github.com/igeargeek/igg-golang-api-utils/utils"
)

type ExampleCreate struct {
	FirstName string `json:"first_name" validate:"required"`
	LastName  uint   `json:"last_name" validate:"required"`
	Email     string `json:"email" validate:"required"`
	Password  string `json:"password" validate:"required"`
}

func ExampleCreateValidator(c *fiber.Ctx) error {
	return ValidateStruct[ExampleCreate](c, nil)
}

### Use in Routes

app.Post("/example", validators.ExampleCreateValidator, exampleController.Create)

```

