package controllers

import (
	"fmt"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"github.com/snowlyg/go-tenancy/common"
	"github.com/snowlyg/go-tenancy/models"
	"github.com/snowlyg/go-tenancy/services"
	"github.com/snowlyg/go-tenancy/sysinit"
	"github.com/snowlyg/go-tenancy/validatas"
)

type RoleController struct {
	Ctx     iris.Context
	Service services.RoleService
}

// GetRoles handles GET: http://localhost:8080/role/table.
func (c *RoleController) GetTable() interface{} {
	args := map[string]interface{}{}
	roles := sysinit.RoleService.GetAll(args, false)

	return common.Table{Code: 0, Msg: "", Count: len(roles), Data: roles}
}

// Get handles GET: http://localhost:8080/role.
func (c *RoleController) Get() mvc.Result {
	return mvc.View{
		Name: "role/index.html",
	}
}

// Get handles GET: http://localhost:8080/role/create.
func (c *RoleController) GetCreate() mvc.Result {
	return mvc.View{
		Name: "role/add.html",
	}
}

// Get handles GET: http://localhost:8080/role/id.
func (c *RoleController) GetBy(id uint) mvc.Result {
	role, _ := c.Service.GetByID(id)
	return mvc.View{
		Name: "role/edit.html",
		Data: iris.Map{
			"Role": role,
		},
	}
}

// Get handles Post: http://localhost:8080/role.
// 使用 ReadJSON 获取数据前端数据需要格式化成json, JSON.stringify(data.field),
func (c *RoleController) Post() interface{} {

	var role models.Role

	if err := c.Ctx.ReadJSON(&role); err != nil {
		return common.ActionResponse{Status: false, Msg: fmt.Sprintf("数据获取错误：%v", err)}
	}

	if err := validatas.Vaild(role); err != nil {
		return common.ActionResponse{Status: false, Msg: fmt.Sprintf("数据验证错误：%v", err)}
	}

	if err := c.Service.Create(&role); err != nil {
		return common.ActionResponse{Status: false, Msg: fmt.Sprintf("用户创建错误：%v", err)}
	}

	return common.ActionResponse{Status: true, Msg: "操作成功"}
}

// Get handles Post: http://localhost:8080/role/id.
func (c *RoleController) PostBy(id uint) interface{} {

	var role models.Role

	if err := c.Ctx.ReadJSON(&role); err != nil {
		return common.ActionResponse{Status: false, Msg: fmt.Sprintf("数据获取错误：%v", err)}
	}

	if err := validatas.Vaild(role); err != nil {
		return common.ActionResponse{Status: false, Msg: fmt.Sprintf("数据验证错误：%v", err)}
	}

	if err := c.Service.Update(id, &role); err != nil {
		return common.ActionResponse{Status: false, Msg: fmt.Sprintf("用户更新错误：%v", err)}
	}

	return common.ActionResponse{Status: true, Msg: "操作成功"}
}

// Get handles Post: http://localhost:8080/role/id.
func (c *RoleController) DeleteBy(id uint) interface{} {
	if err := c.Service.DeleteByID(id); err != nil {
		return common.ActionResponse{Status: false, Msg: fmt.Sprintf("用户删除错误：%v", err)}
	}

	return common.ActionResponse{Status: true, Msg: "操作成功"}
}