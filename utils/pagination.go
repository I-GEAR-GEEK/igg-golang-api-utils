package utils

import (
	"errors"
	"strconv"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/samber/lo"
)

type RespPagination struct {
	Page           int
	PerPage        int
	Offset         int
	OrderField     string
	OrderDirection string
	Search         string
}

type Pagination struct{}

type IPagination interface {
	GetPagination(c *fiber.Ctx, allowOrderField []string) (*RespPagination, error)
}

func NewPagination() IPagination {
	return &Pagination{}
}

func (p *Pagination) GetPagination(c *fiber.Ctx, allowOrderField []string) (*RespPagination, error) {
	page := 1
	if c.Query("page") != "" {
		pageInt, err := strconv.Atoi(c.Query("page"))
		if err != nil {
			return nil, errors.New("Invalid page")
		}
		page = pageInt
	}

	perPage := 10
	if c.Query("per_page") != "" {
		perPageInt, err := strconv.Atoi(c.Query("per_page"))
		if err != nil {
			return nil, errors.New("Invalid perPage")
		}
		perPage = perPageInt
	}

	offset := (page - 1) * perPage

	orderField := "id"
	if c.Query("order_field") != "" {
		if !lo.Contains[string](allowOrderField, c.Query("order_field")) {
			return nil, errors.New("Invalid order field")
		}
		orderField = c.Query("order_field")
	}

	orderDirection := "desc"
	if c.Query("order_direction") != "" {
		if !lo.Contains[string]([]string{"desc", "asc"}, c.Query("order_direction")) {
			return nil, errors.New("Invalid order direction")
		}
		orderDirection = c.Query("order_direction")
	}

	data := &RespPagination{
		Page:           page,
		PerPage:        perPage,
		Offset:         offset,
		OrderField:     orderField,
		OrderDirection: orderDirection,
	}

	return data, nil
}
