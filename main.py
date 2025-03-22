import requests
import time
import json
from dotenv import load_dotenv
import os
from typing import Dict, Any
from dataclasses import dataclass
from datetime import datetime, timezone

# Загрузка переменных окружения
load_dotenv()
TOKEN = os.getenv("TELEGRAM_BOT_TOKEN")
ADMIN_CHAT_ID = os.getenv("ADMIN_CHAT_ID")
TELEGRAM_URL = f"https://api.telegram.org/bot{TOKEN}/"
BACKEND_BASE_URL = 'https://nz1gf6-2a00-1370-81a6-6f51-b11e-a6ed-9bc5-478e.ru.tuna.am'

@dataclass
class UserState:
    state: str
    data: Dict[str, Any]
    last_command: str = ""
    role: str = ""
    name: str = ""
    team_id: int = None
    team_name: str = ""

# Глобальные состояния
user_states: Dict[int, UserState] = {}
welcomed_users = set()

def get_updates(offset=None):
    url = TELEGRAM_URL + "getUpdates"
    params = {
        "timeout": 100,
        "offset": offset
    }
    response = requests.get(url, params=params)
    return response.json()

def send_message(chat_id, text, reply_markup=None, parse_mode="Markdown"):
    url = TELEGRAM_URL + "sendMessage"
    payload = {
        "chat_id": chat_id,
        "text": text,
        "parse_mode": parse_mode
    }
    if reply_markup:
        payload["reply_markup"] = reply_markup
    try:
        response = requests.post(url, json=payload)
        if response.status_code != 200:
            print(f"Ошибка отправки сообщения: {response.text}")
    except Exception as e:
        print(f"Исключение при отправке сообщения: {e}")

def answer_callback(callback_query_id):
    url = TELEGRAM_URL + "answerCallbackQuery"
    payload = {"callback_query_id": callback_query_id}
    requests.post(url, json=payload)

def delete_message(chat_id, message_id):
    url = TELEGRAM_URL + "deleteMessage"
    payload = {"chat_id": chat_id, "message_id": message_id}
    requests.post(url, json=payload)

def auth_get_request(chat_id):
    url = f"{BACKEND_BASE_URL}/auth"
    params = {"telegram_id": str(chat_id)}
    headers = {
        "User-Agent": "TelegramBot/1.0",
        "Accept": "application/json",
    }
    try:
        response = requests.get(url, params=params, headers=headers)
        print(f"Auth response status: {response.status_code}")  # Отладочный вывод
        print(f"Auth response body: {response.text}")  # Отладочный вывод
        
        if response.status_code == 200:
            return {"success": True, "data": response.json()}
        elif response.status_code == 401:
            return {"success": False, "error": "Пользователь не найден"}
        else:
            return {"success": False, "error": f"Неожиданный статус: {response.status_code}"}
    except Exception as e:
        print(f"Auth request exception: {str(e)}")  # Отладочный вывод
        return {"success": False, "error": str(e)}

def auth_post_request(chat_id, name, role):
    url = f"{BACKEND_BASE_URL}/auth"
    payload = {"telegram_id": str(chat_id), "name": name, "role": role}
    headers = {
        "User-Agent": "TelegramBot/1.0",
        "Accept": "application/json",
        "Content-Type": "application/json",
    }
    try:
        print(f"Auth post payload: {payload}")  # Отладочный вывод
        response = requests.post(url, json=payload, headers=headers)
        print(f"Auth post response status: {response.status_code}")  # Отладочный вывод
        print(f"Auth post response body: {response.text}")  # Отладочный вывод
        
        if response.status_code == 200:
            return {"success": True, "data": response.json()}
        else:
            error_msg = response.json().get("error", f"Статус: {response.status_code}")
            return {"success": False, "error": error_msg}
    except Exception as e:
        print(f"Auth post request exception: {str(e)}")  # Отладочный вывод
        return {"success": False, "error": str(e)}

def send_welcome(chat_id):
    keyboard = {
        "inline_keyboard": [
            [{"text": "Начать", "callback_data": "start"}]
        ]
    }
    send_message(chat_id, "👋 Добро пожаловать в систему управления задачами и встречами команды!\n\nНажмите кнопку 'Начать' для регистрации или входа.", reply_markup=keyboard)

def send_role_selection(chat_id, name):
    keyboard = {
        "inline_keyboard": [
            [
                {"text": "👨‍💼 Руководитель", "callback_data": "role_manager"},
                {"text": "👥 Участник", "callback_data": "role_member"}
            ]
        ]
    }
    send_message(chat_id, f"Приятно познакомиться, {name}! Выберите вашу роль:", reply_markup=keyboard)

def send_main_menu(chat_id, user_state: UserState):
    keyboard_buttons = []
    
    if user_state.role == "manager":
        if user_state.team_id:
            # Менеджер с командой
            keyboard_buttons.extend([
                [{"text": "👥 Управление командой", "callback_data": "manage_team"}],
                [{"text": "📋 Управление задачами", "callback_data": "manage_tasks"}],
                [{"text": "📅 Управление встречами", "callback_data": "manage_meetings"}]
            ])
        else:
            # Менеджер без команды
            keyboard_buttons.append([{"text": "📝 Создать команду", "callback_data": "create_team"}])
    else:
        # Участник
        if user_state.team_id:
            # Участник с командой
            keyboard_buttons.extend([
                [{"text": "📋 Мои задачи", "callback_data": "my_tasks"}],
                [{"text": "📅 Мои встречи", "callback_data": "my_meetings"}]
            ])
        else:
            # Участник без команды
            keyboard_buttons.append([{"text": "🤝 Присоединиться к команде", "callback_data": "join_team"}])
    
    keyboard_buttons.append([{"text": "👤 Мой профиль", "callback_data": "my_profile"}])
    
    keyboard = {
        "inline_keyboard": keyboard_buttons
    }
    
    message = f"🏠 *Главное меню*\n"
    if user_state.team_name:
        message += f"\nВаша команда: *{user_state.team_name}*"
    elif user_state.role == "manager":
        message += "\n\n_Создайте свою команду, чтобы начать работу_"
    elif user_state.role == "member":
        message += "\n\n_Присоединитесь к команде, чтобы начать работу_"
    
    print(f"User state: {user_state}")  # Отладочный вывод
    print(f"Sending menu with buttons: {keyboard_buttons}")  # Отладочный вывод
    
    send_message(chat_id, message, reply_markup=keyboard)

def send_profile_menu(chat_id, user_state: UserState):
    message = (
        f"*👤 Профиль*\n\n"
        f"*Имя:* {user_state.name}\n"
        f"*Роль:* {'Руководитель' if user_state.role == 'manager' else 'Участник'}\n"
    )
    
    if user_state.team_name:
        message += f"*Команда:* {user_state.team_name}\n"
        # Добавляем кнопку выхода из команды только для участников
        if user_state.role == "member":
            keyboard = {
                "inline_keyboard": [
                    [{"text": "🚪 Покинуть команду", "callback_data": "leave_team"}],
                    [{"text": "🔙 Назад", "callback_data": "back_to_main"}]
                ]
            }
        else:
            keyboard = {
                "inline_keyboard": [
                    [{"text": "🔙 Назад", "callback_data": "back_to_main"}]
                ]
            }
    else:
        message += "_Нет активной команды_\n"
        keyboard = {
            "inline_keyboard": [
                [{"text": "🔙 Назад", "callback_data": "back_to_main"}]
            ]
        }
    
    send_message(chat_id, message, reply_markup=keyboard)

def process_start_command(chat_id, command_args=None):
    # Проверяем наличие параметра приглашения
    if command_args:
        invite_code = command_args
        auth_result = auth_get_request(chat_id)
        
        if auth_result["success"]:
            # Пользователь уже зарегистрирован
            user_data = auth_result["data"]
            user_state = UserState(
                state="authorized",
                data={},
                role=user_data.get("Role", ""),
                name=user_data.get("Name", ""),
                team_id=user_data.get("TeamID")
            )
            user_states[chat_id] = user_state
            
            # Если пользователь уже в команде
            if user_state.team_id:
                send_message(chat_id, "❌ Вы уже состоите в команде")
                send_main_menu(chat_id, user_state)
                return
            
            # Пробуем присоединить к команде
            result = team_join_request(chat_id, invite_code)
            if result["success"]:
                send_message(chat_id, "✅ Вы успешно присоединились к команде!")
                # Обновляем информацию о команде
                team_info = team_get_my_request(chat_id)
                if team_info["success"]:
                    user_state.team_name = team_info["data"].get("Name", "")
                    user_state.team_id = team_info["data"].get("ID")
                send_main_menu(chat_id, user_state)
            else:
                send_message(chat_id, f"❌ Ошибка: {result['error']}")
                send_main_menu(chat_id, user_state)
        else:
            # Новый пользователь - сохраняем код приглашения
            user_state = UserState(state="awaiting_name", data={"invite_code": invite_code})
            user_states[chat_id] = user_state
            send_message(chat_id, "👋 Давайте познакомимся! Как вас зовут?")
        return

    # Стандартная обработка /start без параметров
    auth_result = auth_get_request(chat_id)
    print(f"Auth result: {auth_result}")
    
    if auth_result["success"]:
        user_data = auth_result["data"]
        print(f"User data: {user_data}")
        
        user_state = UserState(
            state="authorized",
            data={},
            role=user_data.get("Role", ""),
            name=user_data.get("Name", ""),
            team_id=user_data.get("TeamID")
        )
        
        if user_state.team_id:
            team_result = team_get_my_request(chat_id)
            if team_result["success"]:
                user_state.team_name = team_result["data"].get("Name", "")
        
        user_states[chat_id] = user_state
        send_main_menu(chat_id, user_states[chat_id])
    else:
        user_states[chat_id] = UserState(state="awaiting_name", data={})
        send_message(chat_id, "👋 Давайте познакомимся! Как вас зовут?")

def process_callback(callback):
    callback_id = callback["id"]
    chat_id = callback["message"]["chat"]["id"]
    message_id = callback["message"]["message_id"]
    data = callback.get("data")
    
    answer_callback(callback_id)
    delete_message(chat_id, message_id)
    
    user_state = user_states.get(chat_id)
    if not user_state:
        user_state = UserState(state="initial", data={})
        user_states[chat_id] = user_state

    if data == "start":
        process_start_command(chat_id)
        return

    if data.startswith("role_"):
        if user_state.state == "awaiting_role":
            role = data.split("_")[1]
            name = user_state.data.get("name")
            
            result = auth_post_request(chat_id, name, role)
            if result["success"]:
                user_state.role = role
                user_state.name = name
                user_state.state = "authorized"
                send_message(chat_id, "✅ Вы успешно зарегистрированы!")
                
                # Проверяем наличие сохраненного кода приглашения
                invite_code = user_state.data.get("invite_code")
                if invite_code:
                    result = team_join_request(chat_id, invite_code)
                    if result["success"]:
                        send_message(chat_id, "✅ Вы успешно присоединились к команде!")
                        team_info = team_get_my_request(chat_id)
                        if team_info["success"]:
                            user_state.team_name = team_info["data"].get("Name", "")
                            user_state.team_id = team_info["data"].get("ID")
                    else:
                        send_message(chat_id, f"❌ Ошибка присоединения к команде: {result['error']}")
                
                send_main_menu(chat_id, user_state)
            else:
                send_message(chat_id, f"❌ Ошибка регистрации: {result['error']}")
                process_start_command(chat_id)
            return

    # Обработка создания команды
    if data == "create_team":
        if user_state.role == "manager":
            user_state.state = "awaiting_team_name"
            send_message(chat_id, "Введите название команды:")
        else:
            send_message(chat_id, "❌ Только руководитель может создать команду")
            send_main_menu(chat_id, user_state)
        return

    # Добавим обработку команд для команды
    if data == "manage_team":
        send_team_management_menu(chat_id)
    elif data == "join_team":
        send_team_join_menu(chat_id)
    elif data == "team_info":
        result = team_get_my_request(chat_id)
        if result["success"]:
            team_data = result["data"]
            message = (
                f"*Информация о команде*\n"
                f"Название: *{team_data.get('Name', 'Н/Д')}*\n"
                f"Описание: _{team_data.get('Description', 'Отсутствует')}_"
            )
            keyboard = {
                "inline_keyboard": [[{"text": "🔙 Назад", "callback_data": "manage_team"}]]
            }
            send_message(chat_id, message, reply_markup=keyboard)
        else:
            send_message(chat_id, f"❌ Ошибка: {result['error']}")
    elif data == "team_invite":
        result = team_get_invite_request(chat_id)
        if result["success"]:
            invite_link = result["data"].strip('"')  # Удаляем кавычки
            message = f"*Ссылка для приглашения в команду:*\n`{invite_link}`"
            keyboard = {
                "inline_keyboard": [[{"text": "🔙 Назад", "callback_data": "manage_team"}]]
            }
            send_message(chat_id, message, reply_markup=keyboard)
        else:
            send_message(chat_id, f"❌ Ошибка: {result['error']}")
    elif data == "team_members":
        result = team_get_members_request(chat_id)
        if result["success"]:
            members = result["data"]
            message = "*Участники команды:*\n\n"
            keyboard = {"inline_keyboard": []}
            
            for member in members:
                member_name = member.get('Name', 'Н/Д')
                member_id = member.get('TelegramID')
                message += f"👤 {member_name}\n"
                if user_state.role == "manager" and member_id:
                    keyboard["inline_keyboard"].append([{
                        "text": f"❌ Исключить {member_name}",
                        "callback_data": f"kick_member_{member_id}"
                    }])
            
            keyboard["inline_keyboard"].append([{"text": "🔙 Назад", "callback_data": "manage_team"}])
            send_message(chat_id, message, reply_markup=keyboard)
        else:
            send_message(chat_id, f"❌ Ошибка: {result['error']}")
    elif data == "enter_invite_code":
        user_state.state = "awaiting_invite_code"
        send_message(chat_id, "Введите код приглашения:")
    elif data == "back_to_main":
        send_main_menu(chat_id, user_state)
    elif data.startswith("kick_member_"):
        member_id = data.split("_")[2]
        result = team_kick_member_request(chat_id, member_id)
        if result["success"]:
            send_message(chat_id, "✅ Участник успешно исключен из команды")
            # Обновляем список участников
            process_callback({"id": callback_id, "message": callback["message"], "data": "team_members"})
        else:
            send_message(chat_id, f"❌ Ошибка: {result['error']}")
    elif data == "team_delete":
        keyboard = {
            "inline_keyboard": [
                [
                    {"text": "✅ Да, удалить", "callback_data": "confirm_team_delete"},
                    {"text": "❌ Нет, отмена", "callback_data": "manage_team"}
                ]
            ]
        }
        send_message(chat_id, "⚠️ *Вы уверены, что хотите удалить команду?*\nЭто действие нельзя отменить.", reply_markup=keyboard)
    elif data == "confirm_team_delete":
        result = team_delete_request(chat_id)
        if result["success"]:
            send_message(chat_id, "✅ Команда успешно удалена")
            user_state.team_name = ""
            user_state.team_id = None
            send_main_menu(chat_id, user_state)
        else:
            send_message(chat_id, f"❌ Ошибка: {result['error']}")
            send_team_management_menu(chat_id)
    elif data == "my_profile":
        send_profile_menu(chat_id, user_state)
    elif data == "leave_team":
        if user_state.role == "member":
            keyboard = {
                "inline_keyboard": [
                    [
                        {"text": "✅ Да, покинуть", "callback_data": "confirm_leave_team"},
                        {"text": "❌ Нет, остаться", "callback_data": "my_profile"}
                    ]
                ]
            }
            send_message(chat_id, "⚠️ *Вы уверены, что хотите покинуть команду?*", reply_markup=keyboard)
        else:
            send_message(chat_id, "❌ Руководитель не может покинуть команду")
            send_profile_menu(chat_id, user_state)
    elif data == "confirm_leave_team":
        result = team_leave_request(chat_id)
        if result["success"]:
            send_message(chat_id, "✅ Вы успешно покинули команду")
            user_state.team_id = None
            user_state.team_name = ""
            send_main_menu(chat_id, user_state)
        else:
            send_message(chat_id, f"❌ Ошибка: {result['error']}")
            send_profile_menu(chat_id, user_state)
    elif data == "manage_tasks":
        send_task_management_menu(chat_id, user_state)
    elif data == "my_tasks":
        send_my_tasks_menu(chat_id)
    elif data == "issued_tasks":
        send_issued_tasks_menu(chat_id)
    elif data == "create_task":
        user_state.state = "awaiting_task_type"
        keyboard = {
            "inline_keyboard": [
                [
                    {"text": "👥 Командная задача", "callback_data": "create_team_task"},
                    {"text": "👤 Персональная задача", "callback_data": "create_personal_task"}
                ],
                [{"text": "🔙 Назад", "callback_data": "manage_tasks"}]
            ]
        }
        send_message(chat_id, "Выберите тип задачи:", reply_markup=keyboard)
    elif data.startswith("create_team_task") or data.startswith("create_personal_task"):
        user_state.data["task_type"] = "team" if data.startswith("create_team_task") else "personal"
        user_state.state = "awaiting_task_title"
        send_message(chat_id, "Введите название задачи:")
    elif data.startswith("update_task_status_"):
        task_id = data.split("_")[-1]
        user_state.data["updating_task_id"] = task_id
        keyboard = {
            "inline_keyboard": [
                [
                    {"text": "🔄 В работе", "callback_data": f"set_status_in_progress_{task_id}"},
                    {"text": "✅ Завершена", "callback_data": f"set_status_completed_{task_id}"}
                ],
                [{"text": "🔙 Назад", "callback_data": "my_tasks"}]
            ]
        }
        send_message(chat_id, "Выберите новый статус задачи:", reply_markup=keyboard)
    elif data.startswith("set_status_"):
        parts = data.split("_")
        status = parts[-2]
        task_id = parts[-1]
        user_state.data["updating_task_id"] = task_id
        user_state.data["new_status"] = status
        user_state.state = "awaiting_completion_text"
        if status == "completed":
            send_message(chat_id, "Введите отчет о выполнении задачи (или отправьте '-' если отчет не требуется):")
        else:
            result = tasks_update_status_request(chat_id, task_id, status)
            if result["success"]:
                send_message(chat_id, "✅ Статус задачи обновлен")
                send_my_tasks_menu(chat_id)
            else:
                send_message(chat_id, f"❌ Ошибка обновления статуса: {result['error']}")
    elif data.startswith("delete_task_"):
        task_id = data.split("_")[-1]
        keyboard = {
            "inline_keyboard": [
                [
                    {"text": "✅ Да, удалить", "callback_data": f"confirm_delete_task_{task_id}"},
                    {"text": "❌ Нет, отмена", "callback_data": "issued_tasks"}
                ]
            ]
        }
        send_message(chat_id, "⚠️ Вы уверены, что хотите удалить эту задачу?", reply_markup=keyboard)
    elif data.startswith("confirm_delete_task_"):
        task_id = data.split("_")[-1]
        result = tasks_delete_request(chat_id, task_id)
        if result["success"]:
            send_message(chat_id, "✅ Задача успешно удалена")
            send_issued_tasks_menu(chat_id)
        else:
            send_message(chat_id, f"❌ Ошибка удаления задачи: {result['error']}")
            send_issued_tasks_menu(chat_id)
    elif data.startswith("assign_task_"):
        assigned_to = data.split("_")[2]
        result = tasks_create_request(
            chat_id,
            user_state.data["task_title"],
            user_state.data["task_description"],
            user_state.data["task_deadline"],
            is_team=False,
            assigned_to=assigned_to
        )
        if result["success"]:
            send_message(chat_id, "✅ Задача успешно создана")
            send_task_management_menu(chat_id, user_state)
        else:
            send_message(chat_id, f"❌ Ошибка создания задачи: {result['error']}")
            send_task_management_menu(chat_id, user_state)
        user_state.state = "authorized"

def process_message(message):
    chat_id = message["chat"]["id"]
    text = message.get("text", "")
    
    if text.startswith('/start'):
        # Проверяем наличие параметра в команде start
        parts = text.split()
        if len(parts) > 1:
            process_start_command(chat_id, parts[1])
        else:
            process_start_command(chat_id)
        return

    user_state = user_states.get(chat_id)
    if not user_state:
        user_state = UserState(state="initial", data={})
        user_states[chat_id] = user_state

    if text.startswith('/'):
        if text == '/start':
            process_start_command(chat_id)
        elif text == '/menu':
            if user_state.state == "authorized":
                send_main_menu(chat_id, user_state)
            else:
                process_start_command(chat_id)
        return

    if user_state.state == "awaiting_name":
        user_state.data["name"] = text
        user_state.state = "awaiting_role"
        send_role_selection(chat_id, text)
        return

    if user_state.state == "awaiting_team_name":
        user_state.data["team_name"] = text
        user_state.state = "awaiting_team_description"
        send_message(chat_id, "Введите описание команды (или отправьте '-' если описание не требуется):")
        return

    elif user_state.state == "awaiting_team_description":
        description = text if text != "-" else ""
        result = team_create_request(chat_id, user_state.data["team_name"], description)
        if result["success"]:
            send_message(chat_id, "✅ Команда успешно создана!")
            # Обновляем состояние пользователя
            user_state.team_name = user_state.data["team_name"]
            user_state.team_id = result["data"].get("ID")  # Используем 'ID' вместо 'id'
            user_state.state = "authorized"
            send_main_menu(chat_id, user_state)
        else:
            send_message(chat_id, f"❌ Ошибка создания команды: {result['error']}")
            send_main_menu(chat_id, user_state)
        return

    elif user_state.state == "awaiting_invite_code":
        result = team_join_request(chat_id, text)
        if result["success"]:
            send_message(chat_id, "✅ Вы успешно присоединились к команде!")
            # Обновляем информацию о команде
            team_info = team_get_my_request(chat_id)
            if team_info["success"]:
                user_state.team_name = team_info["data"].get("Name", "")  # Используем 'Name' вместо 'name'
                user_state.team_id = team_info["data"].get("ID")  # Используем 'ID' вместо 'id'
            user_state.state = "authorized"
            send_main_menu(chat_id, user_state)
        else:
            send_message(chat_id, f"❌ Ошибка: {result['error']}")
            send_team_join_menu(chat_id)
        return

    elif user_state.state == "awaiting_task_title":
        user_state.data["task_title"] = text
        user_state.state = "awaiting_task_description"
        send_message(chat_id, "Введите описание задачи:")
    elif user_state.state == "awaiting_task_description":
        user_state.data["task_description"] = text
        user_state.state = "awaiting_task_deadline"
        send_message(chat_id, "Введите срок выполнения задачи в формате ГГГГ-ММ-ДД ЧЧ:ММ\nНапример: 2024-03-25 15:00")
        return
    elif user_state.state == "awaiting_task_deadline":
        try:
            # Парсим введенную дату
            deadline = datetime.strptime(text, "%Y-%m-%d %H:%M")
            # Добавляем часовой пояс (UTC)
            deadline = deadline.replace(tzinfo=timezone.utc)
            user_state.data["task_deadline"] = deadline.isoformat()
            
            if user_state.data["task_type"] == "personal":
                user_state.state = "awaiting_task_assignee"
                result = team_get_members_request(chat_id)
                if result["success"]:
                    members = result["data"]
                    keyboard = {"inline_keyboard": []}
                    for member in members:
                        keyboard["inline_keyboard"].append([{
                            "text": member.get("Name", "Н/Д"),
                            "callback_data": f"assign_task_{member.get('TelegramID')}"
                        }])
                    keyboard["inline_keyboard"].append([{"text": "🔙 Отмена", "callback_data": "manage_tasks"}])
                    send_message(chat_id, "Выберите исполнителя задачи:", reply_markup=keyboard)
                else:
                    send_message(chat_id, f"❌ Ошибка получения списка участников: {result['error']}")
                    send_task_management_menu(chat_id, user_state)
            else:
                # Создаем командную задачу
                result = tasks_create_request(
                    chat_id,
                    user_state.data["task_title"],
                    user_state.data["task_description"],
                    user_state.data["task_deadline"],
                    is_team=True
                )
                if result["success"]:
                    send_message(chat_id, "✅ Задача успешно создана")
                    send_task_management_menu(chat_id, user_state)
                else:
                    send_message(chat_id, f"❌ Ошибка создания задачи: {result['error']}")
                    send_task_management_menu(chat_id, user_state)
        except ValueError:
            send_message(chat_id, "❌ Неверный формат даты. Попробуйте еще раз.\nФормат: ГГГГ-ММ-ДД ЧЧ:ММ")
    elif user_state.state == "awaiting_completion_text":
        completion_text = None if text == "-" else text
        result = tasks_update_status_request(
            chat_id,
            user_state.data["updating_task_id"],
            user_state.data["new_status"],
            completion_text=completion_text
        )
        if result["success"]:
            send_message(chat_id, "✅ Статус задачи обновлен")
            send_my_tasks_menu(chat_id)
        else:
            send_message(chat_id, f"❌ Ошибка обновления статуса: {result['error']}")
        user_state.state = "authorized"

    # Отправляем сообщение о навигации только если пользователь находится в состоянии "authorized"
    if user_state.state == "authorized":
        send_message(chat_id, "Используйте кнопки меню для навигации или команду /menu для вызова главного меню.")

def team_create_request(chat_id, name, description):
    url = f"{BACKEND_BASE_URL}/team"
    params = {"telegram_id": str(chat_id)}
    payload = {"name": name, "description": description}
    headers = {
        "User-Agent": "TelegramBot/1.0",
        "Accept": "application/json",
        "Content-Type": "application/json",
    }
    try:
        response = requests.post(url, params=params, json=payload, headers=headers)
        if response.status_code == 200:
            return {"success": True, "data": response.json()}
        else:
            return {"success": False, "error": response.json().get("error", f"Статус: {response.status_code}")}
    except Exception as e:
        return {"success": False, "error": str(e)}

def team_join_request(chat_id, invite_code):
    url = f"{BACKEND_BASE_URL}/team/join"
    payload = {"telegram_id": str(chat_id), "invite_code": invite_code}
    headers = {
        "User-Agent": "TelegramBot/1.0",
        "Accept": "application/json",
        "Content-Type": "application/json",
    }
    try:
        response = requests.post(url, json=payload, headers=headers)
        if response.status_code == 200:
            return {"success": True, "data": response.json()}
        else:
            return {"success": False, "error": response.json().get("error", f"Статус: {response.status_code}")}
    except Exception as e:
        return {"success": False, "error": str(e)}

def team_get_my_request(chat_id):
    url = f"{BACKEND_BASE_URL}/team/my"
    params = {"telegram_id": str(chat_id)}
    headers = {
        "User-Agent": "TelegramBot/1.0",
        "Accept": "application/json",
    }
    try:
        response = requests.get(url, params=params, headers=headers)
        if response.status_code == 200:
            return {"success": True, "data": response.json()}
        else:
            return {"success": False, "error": response.json().get("error", f"Статус: {response.status_code}")}
    except Exception as e:
        return {"success": False, "error": str(e)}

def team_get_invite_request(chat_id):
    url = f"{BACKEND_BASE_URL}/team/invite"
    params = {"telegram_id": str(chat_id)}
    headers = {
        "User-Agent": "TelegramBot/1.0",
        "Accept": "application/json",
    }
    try:
        response = requests.get(url, params=params, headers=headers)
        if response.status_code == 200:
            return {"success": True, "data": response.text}  # Возвращает URL для приглашения
        else:
            return {"success": False, "error": response.json().get("error", f"Статус: {response.status_code}")}
    except Exception as e:
        return {"success": False, "error": str(e)}

def team_get_members_request(chat_id):
    url = f"{BACKEND_BASE_URL}/team/members"
    params = {"telegram_id": str(chat_id)}
    headers = {
        "User-Agent": "TelegramBot/1.0",
        "Accept": "application/json",
    }
    try:
        response = requests.get(url, params=params, headers=headers)
        if response.status_code == 200:
            return {"success": True, "data": response.json()}
        else:
            return {"success": False, "error": response.json().get("error", f"Статус: {response.status_code}")}
    except Exception as e:
        return {"success": False, "error": str(e)}

def team_kick_member_request(chat_id, kick_telegram_id):
    url = f"{BACKEND_BASE_URL}/team/kick"
    params = {
        "telegram_id": str(chat_id),
        "kick_telegram_id": str(kick_telegram_id)
    }
    headers = {
        "User-Agent": "TelegramBot/1.0",
        "Accept": "application/json",
    }
    try:
        response = requests.get(url, params=params, headers=headers)
        if response.status_code == 200:
            return {"success": True, "data": response.json()}
        else:
            return {"success": False, "error": response.json().get("error", f"Статус: {response.status_code}")}
    except Exception as e:
        return {"success": False, "error": str(e)}

def team_leave_request(chat_id):
    url = f"{BACKEND_BASE_URL}/team/leave"
    params = {"telegram_id": str(chat_id)}
    headers = {
        "User-Agent": "TelegramBot/1.0",
        "Accept": "application/json",
    }
    try:
        response = requests.get(url, params=params, headers=headers)
        if response.status_code == 200:
            return {"success": True, "data": response.json()}
        else:
            return {"success": False, "error": response.json().get("error", f"Статус: {response.status_code}")}
    except Exception as e:
        return {"success": False, "error": str(e)}

# Добавим функции для отображения меню команды
def send_team_management_menu(chat_id):
    keyboard = {
        "inline_keyboard": [
            [{"text": "📋 Информация о команде", "callback_data": "team_info"}],
            [{"text": "🔗 Получить ссылку-приглашение", "callback_data": "team_invite"}],
            [{"text": "👥 Список участников", "callback_data": "team_members"}],
            [{"text": "❌ Удалить команду", "callback_data": "team_delete"}],
            [{"text": "🔙 Назад", "callback_data": "back_to_main"}]
        ]
    }
    send_message(chat_id, "*Управление командой*\nВыберите действие:", reply_markup=keyboard)

def send_team_join_menu(chat_id):
    keyboard = {
        "inline_keyboard": [
            [{"text": "🔗 Ввести код приглашения", "callback_data": "enter_invite_code"}],
            [{"text": "🔙 Назад", "callback_data": "back_to_main"}]
        ]
    }
    send_message(chat_id, "*Присоединение к команде*\nВыберите действие:", reply_markup=keyboard)

def team_delete_request(chat_id):
    url = f"{BACKEND_BASE_URL}/team"
    params = {"telegram_id": str(chat_id)}
    headers = {
        "User-Agent": "TelegramBot/1.0",
        "Accept": "application/json",
    }
    try:
        response = requests.delete(url, params=params, headers=headers)
        if response.status_code == 200:
            return {"success": True, "data": response.json()}
        else:
            return {"success": False, "error": response.json().get("error", f"Статус: {response.status_code}")}
    except Exception as e:
        return {"success": False, "error": str(e)}

def tasks_create_request(chat_id, title, description, deadline, is_team=False, assigned_to=None):
    url = f"{BACKEND_BASE_URL}/tasks"
    params = {"telegram_id": str(chat_id)}
    payload = {
        "title": title,
        "description": description,
        "deadline": deadline,
        "is_team": is_team
    }
    if assigned_to:
        payload["assigned_to"] = assigned_to

    headers = {
        "User-Agent": "TelegramBot/1.0",
        "Accept": "application/json",
        "Content-Type": "application/json",
    }
    try:
        response = requests.post(url, params=params, json=payload, headers=headers)
        if response.status_code == 200:
            return {"success": True, "data": response.json()}
        else:
            return {"success": False, "error": response.json().get("error", f"Статус: {response.status_code}")}
    except Exception as e:
        return {"success": False, "error": str(e)}

def tasks_get_request(chat_id):
    url = f"{BACKEND_BASE_URL}/tasks"
    params = {"telegram_id": str(chat_id)}
    headers = {
        "User-Agent": "TelegramBot/1.0",
        "Accept": "application/json",
    }
    try:
        response = requests.get(url, params=params, headers=headers)
        if response.status_code == 200:
            return {"success": True, "data": response.json()}
        else:
            return {"success": False, "error": response.json().get("error", f"Статус: {response.status_code}")}
    except Exception as e:
        return {"success": False, "error": str(e)}

def tasks_get_issued_request(chat_id):
    url = f"{BACKEND_BASE_URL}/tasks/issued"
    params = {"telegram_id": str(chat_id)}
    headers = {
        "User-Agent": "TelegramBot/1.0",
        "Accept": "application/json",
    }
    try:
        response = requests.get(url, params=params, headers=headers)
        if response.status_code == 200:
            return {"success": True, "data": response.json()}
        else:
            return {"success": False, "error": response.json().get("error", f"Статус: {response.status_code}")}
    except Exception as e:
        return {"success": False, "error": str(e)}

def tasks_delete_request(chat_id, task_id):
    url = f"{BACKEND_BASE_URL}/tasks/{task_id}"
    params = {"telegram_id": str(chat_id)}
    headers = {
        "User-Agent": "TelegramBot/1.0",
        "Accept": "application/json",
    }
    try:
        response = requests.delete(url, params=params, headers=headers)
        if response.status_code == 200:
            return {"success": True, "data": response.json()}
        else:
            return {"success": False, "error": response.json().get("error", f"Статус: {response.status_code}")}
    except Exception as e:
        return {"success": False, "error": str(e)}

def tasks_update_status_request(chat_id, task_id, status, completion_text=None, attachment=None):
    url = f"{BACKEND_BASE_URL}/tasks/{task_id}/status"
    params = {"telegram_id": str(chat_id)}
    payload = {"status": status}
    if completion_text:
        payload["completion_text"] = completion_text
    if attachment:
        payload["attachment"] = attachment

    headers = {
        "User-Agent": "TelegramBot/1.0",
        "Accept": "application/json",
        "Content-Type": "application/json",
    }
    try:
        response = requests.put(url, params=params, json=payload, headers=headers)
        if response.status_code == 200:
            return {"success": True, "data": response.json()}
        else:
            return {"success": False, "error": response.json().get("error", f"Статус: {response.status_code}")}
    except Exception as e:
        return {"success": False, "error": str(e)}

def format_task_status(status):
    status_emojis = {
        "assigned": "📝",
        "in_progress": "🔄",
        "completed": "✅"
    }
    status_names = {
        "assigned": "Назначена",
        "in_progress": "В работе",
        "completed": "Завершена"
    }
    return f"{status_emojis.get(status, '❓')} {status_names.get(status, status)}"

def format_task_deadline(deadline_str):
    try:
        deadline = datetime.fromisoformat(deadline_str.replace('Z', '+00:00'))
        now = datetime.now(deadline.tzinfo)
        days_left = (deadline - now).days
        
        if days_left < 0:
            return "⌛️ Срок истек"
        elif days_left == 0:
            return f"⏳ Сегодня в {deadline.strftime('%H:%M')}"
        elif days_left == 1:
            return "⏳ Завтра"
        else:
            return f"📅 Через {days_left} дней"
    except:
        return "❓ Неизвестный срок"

def send_task_management_menu(chat_id, user_state: UserState):
    keyboard = {
        "inline_keyboard": [
            [{"text": "📝 Создать задачу", "callback_data": "create_task"}],
            [{"text": "📋 Выданные задачи", "callback_data": "issued_tasks"}],
            [{"text": "🔙 Назад", "callback_data": "back_to_main"}]
        ]
    }
    send_message(chat_id, "*Управление задачами*\nВыберите действие:", reply_markup=keyboard)

def send_my_tasks_menu(chat_id):
    result = tasks_get_request(chat_id)
    if not result["success"]:
        send_message(chat_id, f"❌ Ошибка получения задач: {result['error']}")
        return

    tasks = result["data"]
    if not tasks:
        keyboard = {
            "inline_keyboard": [[{"text": "🔙 Назад", "callback_data": "back_to_main"}]]
        }
        send_message(chat_id, "*Мои задачи*\n\n_У вас пока нет задач_", reply_markup=keyboard)
        return

    message = "*Мои задачи:*\n\n"
    keyboard = {"inline_keyboard": []}

    for task in tasks:
        status = format_task_status(task.get("status", "unknown"))
        deadline = format_task_deadline(task.get("deadline", ""))
        message += f"*{task.get('title', 'Без названия')}*\n"
        message += f"_{task.get('description', 'Без описания')}_\n"
        message += f"Статус: {status}\n"
        message += f"Срок: {deadline}\n\n"
        
        if task.get("status") != "completed":
            keyboard["inline_keyboard"].append([{
                "text": f"✏️ Обновить статус: {task.get('title')}",
                "callback_data": f"update_task_status_{task.get('id')}"
            }])

    keyboard["inline_keyboard"].append([{"text": "🔙 Назад", "callback_data": "back_to_main"}])
    send_message(chat_id, message, reply_markup=keyboard)

def send_issued_tasks_menu(chat_id):
    result = tasks_get_issued_request(chat_id)
    if not result["success"]:
        send_message(chat_id, f"❌ Ошибка получения выданных задач: {result['error']}")
        return

    tasks = result["data"]
    if not tasks:
        keyboard = {
            "inline_keyboard": [
                [{"text": "📝 Создать задачу", "callback_data": "create_task"}],
                [{"text": "🔙 Назад", "callback_data": "manage_tasks"}]
            ]
        }
        send_message(chat_id, "*Выданные задачи*\n\n_У вас пока нет выданных задач_", reply_markup=keyboard)
        return

    message = "*Выданные задачи:*\n\n"
    keyboard = {"inline_keyboard": []}

    for task in tasks:
        status = format_task_status(task.get("status", "unknown"))
        deadline = format_task_deadline(task.get("deadline", ""))
        assigned_to = task.get("assigned_to", "Команда" if task.get("is_team") else "Н/Д")
        
        message += f"*{task.get('title', 'Без названия')}*\n"
        message += f"_{task.get('description', 'Без описания')}_\n"
        message += f"Кому: {assigned_to}\n"
        message += f"Статус: {status}\n"
        message += f"Срок: {deadline}\n\n"
        
        keyboard["inline_keyboard"].append([{
            "text": f"❌ Удалить: {task.get('title')}",
            "callback_data": f"delete_task_{task.get('id')}"
        }])

    keyboard["inline_keyboard"].extend([
        [{"text": "📝 Создать задачу", "callback_data": "create_task"}],
        [{"text": "🔙 Назад", "callback_data": "manage_tasks"}]
    ])
    
    send_message(chat_id, message, reply_markup=keyboard)

def main():
    print("Бот запущен и ожидает сообщений...")
    offset = None

    while True:
        try:
            updates = get_updates(offset)
            if updates.get("ok"):
                for update in updates["result"]:
                    offset = update["update_id"] + 1

                    if "callback_query" in update:
                        process_callback(update["callback_query"])
                    elif "message" in update:
                        process_message(update["message"])

        except Exception as e:
            print(f"Ошибка в главном цикле: {e}")
            time.sleep(5)
            continue

        time.sleep(1)

if __name__ == '__main__':
    main()
