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
        "/auth": {
            "get": {
                "description": "Проверяет, зарегистрирован ли пользователь по telegram_id. Если пользователь найден, возвращает его данные, иначе – сообщение об ошибке.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "auth"
                ],
                "summary": "Проверка авторизации пользователя",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Уникальный идентификатор Telegram",
                        "name": "telegram_id",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Данные пользователя",
                        "schema": {
                            "$ref": "#/definitions/auth.RegisterInput"
                        }
                    },
                    "400": {
                        "description": "Ошибка telegram_id is required",
                        "schema": {
                            "$ref": "#/definitions/response.ErrorResponse"
                        }
                    },
                    "401": {
                        "description": "Пользователь не найден",
                        "schema": {
                            "$ref": "#/definitions/response.ErrorResponse"
                        }
                    }
                }
            },
            "post": {
                "description": "Регистрация пользователя с помощью уникального telegram_id",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "auth"
                ],
                "summary": "Регистрация пользователя",
                "parameters": [
                    {
                        "description": "Данные пользователя для регистрации",
                        "name": "input",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/auth.RegisterInput"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Успешная регистрация",
                        "schema": {
                            "$ref": "#/definitions/response.SuccessResponse"
                        }
                    },
                    "400": {
                        "description": "Ошибка валидации",
                        "schema": {
                            "$ref": "#/definitions/response.ErrorResponse"
                        }
                    },
                    "409": {
                        "description": "Пользователь уже зарегистрирован",
                        "schema": {
                            "$ref": "#/definitions/response.ErrorResponse"
                        }
                    },
                    "507": {
                        "description": "Не удалось создать пользователя",
                        "schema": {
                            "$ref": "#/definitions/response.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/team": {
            "post": {
                "description": "Создает команду, если запрос исходит от пользователя с ролью manager.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "team"
                ],
                "summary": "Создание команды",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Уникальный идентификатор Telegram",
                        "name": "telegram_id",
                        "in": "query",
                        "required": true
                    },
                    {
                        "description": "Данные команды",
                        "name": "input",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/team.CreateTeamInput"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Информация о созданной команде",
                        "schema": {
                            "$ref": "#/definitions/response.TeamResponse"
                        }
                    },
                    "400": {
                        "description": "Ошибка валидации или отсутствует telegram_id",
                        "schema": {
                            "$ref": "#/definitions/response.ErrorResponse"
                        }
                    },
                    "403": {
                        "description": "Доступ запрещен (не менеджер)",
                        "schema": {
                            "$ref": "#/definitions/response.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Ошибка создания команды",
                        "schema": {
                            "$ref": "#/definitions/response.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/team/join": {
            "post": {
                "description": "Позволяет пользователю присоединиться к команде, используя пригласительный код.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "team"
                ],
                "summary": "Присоединение к команде",
                "parameters": [
                    {
                        "description": "Данные для присоединения к команде",
                        "name": "input",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/team.InviteJoinRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Успешное присоединение к команде",
                        "schema": {
                            "$ref": "#/definitions/response.SuccessResponse"
                        }
                    },
                    "400": {
                        "description": "Ошибка валидации",
                        "schema": {
                            "$ref": "#/definitions/response.ErrorResponse"
                        }
                    },
                    "404": {
                        "description": "Неверный код приглашения или пользователь не найден",
                        "schema": {
                            "$ref": "#/definitions/response.ErrorResponse"
                        }
                    },
                    "409": {
                        "description": "Вы уже присоединились к этой команде",
                        "schema": {
                            "$ref": "#/definitions/response.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Ошибка при присоединении к команде",
                        "schema": {
                            "$ref": "#/definitions/response.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/team/my": {
            "get": {
                "description": "Возвращает данные о команде, к которой принадлежит пользователь",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "team"
                ],
                "summary": "Получение информации о своей команде",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Уникальный идентификатор Telegram",
                        "name": "telegram_id",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Информация о команде",
                        "schema": {
                            "$ref": "#/definitions/response.TeamResponse"
                        }
                    },
                    "400": {
                        "description": "Отсутствует telegram_id",
                        "schema": {
                            "$ref": "#/definitions/response.ErrorResponse"
                        }
                    },
                    "401": {
                        "description": "Пользователь не найден",
                        "schema": {
                            "$ref": "#/definitions/response.ErrorResponse"
                        }
                    },
                    "404": {
                        "description": "Error:Отсутствует команда у пользователя Сode:USER_HAS_NO_TEAM, Error: Команда не найдена Сode:TEAM_NOT_FOUND",
                        "schema": {
                            "$ref": "#/definitions/response.ErrorCodeResponse"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "auth.RegisterInput": {
            "type": "object",
            "required": [
                "name",
                "role",
                "telegram_id"
            ],
            "properties": {
                "name": {
                    "type": "string"
                },
                "role": {
                    "description": "\"manager\" или \"member\"",
                    "type": "string",
                    "enum": [
                        "manager",
                        "member"
                    ]
                },
                "telegram_id": {
                    "type": "string"
                }
            }
        },
        "response.ErrorCodeResponse": {
            "type": "object",
            "properties": {
                "code": {
                    "type": "string"
                },
                "error": {
                    "type": "string"
                }
            }
        },
        "response.ErrorResponse": {
            "type": "object",
            "properties": {
                "error": {
                    "type": "string"
                }
            }
        },
        "response.SuccessResponse": {
            "type": "object",
            "properties": {
                "message": {
                    "type": "string"
                }
            }
        },
        "response.TeamResponse": {
            "type": "object",
            "properties": {
                "description": {
                    "type": "string"
                },
                "invitelink": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                }
            }
        },
        "team.CreateTeamInput": {
            "type": "object",
            "required": [
                "name"
            ],
            "properties": {
                "description": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                }
            }
        },
        "team.InviteJoinRequest": {
            "type": "object",
            "required": [
                "invite_code",
                "telegram_id"
            ],
            "properties": {
                "invite_code": {
                    "type": "string"
                },
                "telegram_id": {
                    "type": "string"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "",
	Host:             "",
	BasePath:         "",
	Schemes:          []string{},
	Title:            "Сервис для контроля задачами и встречами команды",
	Description:      "",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
