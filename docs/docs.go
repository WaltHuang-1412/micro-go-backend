// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/login": {
            "post": {
                "description": "輸入 email 與密碼後登入並取得 JWT Token",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Auth"
                ],
                "summary": "使用者登入",
                "parameters": [
                    {
                        "description": "登入資訊",
                        "name": "login",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.UserLoginInput"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            }
        },
        "/plans/sections": {
            "get": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "依照排序列出所有區塊",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Plans"
                ],
                "summary": "取得所有區塊（Section）",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/models.Section"
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            },
            "post": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "建立一個新的區塊（自動補上 sort_order）",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Plans"
                ],
                "summary": "建立新區塊（Section）",
                "parameters": [
                    {
                        "description": "區塊資料",
                        "name": "section",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.CreateSectionInput"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "object",
                            "additionalProperties": true
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            }
        },
        "/plans/sections-with-tasks": {
            "get": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "回傳每個區塊與其所屬任務（僅限本人），依照排序排列",
                "tags": [
                    "Plans"
                ],
                "summary": "取得所有區塊（含任務）",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/models.SectionWithTasks"
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            },
            "put": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "依據傳入資料更新 sections 與 tasks 的 sort_order（title/content 不會變動）",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Plans"
                ],
                "summary": "批次更新區塊與任務排序",
                "parameters": [
                    {
                        "description": "排序資料",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/models.SectionWithTasks"
                            }
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            }
        },
        "/plans/sections/{id}": {
            "put": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "根據 ID 修改區塊的標題，僅限本人操作",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Plans"
                ],
                "summary": "更新區塊（Section 標題）",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "Section ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "更新資料",
                        "name": "section",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.UpdateSectionInput"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "object",
                            "additionalProperties": true
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            },
            "delete": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "根據 ID 刪除一個區塊，並重新排序該使用者的其他區塊",
                "tags": [
                    "Plans"
                ],
                "summary": "刪除區塊（Section）",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "Section ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            }
        },
        "/plans/tasks": {
            "post": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "建立新的任務，並自動排序",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Plans"
                ],
                "summary": "建立任務（Task）",
                "parameters": [
                    {
                        "description": "任務內容",
                        "name": "task",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.CreateTaskInput"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "object",
                            "additionalProperties": true
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            }
        },
        "/plans/tasks/{id}": {
            "put": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "根據 ID 更新任務內容",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Plans"
                ],
                "summary": "更新任務（Task）",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "任務 ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "更新資料",
                        "name": "task",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.UpdateTaskInput"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    },
                    "403": {
                        "description": "Forbidden",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            },
            "delete": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "根據 ID 刪除任務，並重新排序同區塊內的任務",
                "tags": [
                    "Plans"
                ],
                "summary": "刪除任務（Task）",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "任務 ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    },
                    "403": {
                        "description": "Forbidden",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            }
        },
        "/profile": {
            "get": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "使用 JWT 取得當前登入者資訊",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "取得個人資訊",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "object",
                            "additionalProperties": true
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            }
        },
        "/register": {
            "post": {
                "description": "使用者註冊帳號",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Auth"
                ],
                "summary": "註冊使用者",
                "parameters": [
                    {
                        "description": "使用者資料",
                        "name": "user",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.UserRegisterInput"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "models.CreateSectionInput": {
            "type": "object",
            "required": [
                "title"
            ],
            "properties": {
                "title": {
                    "type": "string"
                }
            }
        },
        "models.CreateTaskInput": {
            "type": "object",
            "required": [
                "content",
                "section_id",
                "title"
            ],
            "properties": {
                "content": {
                    "type": "string"
                },
                "is_completed": {
                    "type": "boolean"
                },
                "section_id": {
                    "type": "integer"
                },
                "title": {
                    "type": "string"
                }
            }
        },
        "models.Section": {
            "type": "object",
            "properties": {
                "created_at": {
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "sort_order": {
                    "type": "integer"
                },
                "title": {
                    "type": "string"
                },
                "updated_at": {
                    "type": "string"
                }
            }
        },
        "models.SectionWithTasks": {
            "type": "object",
            "properties": {
                "created_at": {
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "sort_order": {
                    "type": "integer"
                },
                "tasks": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/models.Task"
                    }
                },
                "title": {
                    "type": "string"
                },
                "updated_at": {
                    "type": "string"
                }
            }
        },
        "models.Task": {
            "type": "object",
            "properties": {
                "content": {
                    "type": "string"
                },
                "created_at": {
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "is_completed": {
                    "type": "boolean"
                },
                "section_id": {
                    "type": "integer"
                },
                "sort_order": {
                    "type": "integer"
                },
                "title": {
                    "type": "string"
                },
                "updated_at": {
                    "type": "string"
                }
            }
        },
        "models.UpdateSectionInput": {
            "type": "object",
            "required": [
                "title"
            ],
            "properties": {
                "title": {
                    "type": "string"
                }
            }
        },
        "models.UpdateTaskInput": {
            "type": "object",
            "properties": {
                "content": {
                    "type": "string"
                },
                "is_completed": {
                    "type": "boolean"
                },
                "title": {
                    "type": "string"
                }
            }
        },
        "models.UserLoginInput": {
            "type": "object",
            "properties": {
                "email": {
                    "type": "string",
                    "example": "w@w.com"
                },
                "password": {
                    "type": "string",
                    "example": "123456"
                }
            }
        },
        "models.UserRegisterInput": {
            "type": "object",
            "properties": {
                "email": {
                    "type": "string",
                    "example": "w@w.com"
                },
                "password": {
                    "type": "string",
                    "example": "123456"
                },
                "username": {
                    "type": "string",
                    "example": "walter"
                }
            }
        }
    },
    "securityDefinitions": {
        "BearerAuth": {
            "type": "apiKey",
            "name": "Authorization",
            "in": "header"
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "localhost:8088",
	BasePath:         "/api/v1",
	Schemes:          []string{},
	Title:            "Micro Backend API",
	Description:      "使用 JWT 的簡易用戶驗證 API",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
