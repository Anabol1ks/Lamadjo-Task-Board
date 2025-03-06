import requests
import time
import json
from dotenv import load_dotenv
import os

# Загрузка переменных окружения
load_dotenv()
TOKEN = os.getenv("TELEGRAM_BOT_TOKEN")
ADMIN_CHAT_ID = os.getenv("ADMIN_CHAT_ID")  # Задайте chat_id, которому бот должен сразу отправить сообщение при запуске
TELEGRAM_URL = f"https://api.telegram.org/bot{TOKEN}/"

# Базовый URL вашего бэкенда (без завершающего слэша)
BACKEND_BASE_URL = 'https://8b2b28-213-87-86-243.ru.tuna.am'

user_states = {}

# Словарь для хранения зарегистрированных пользователей: chat_id -> role
registered_users = {}

# Множество для отслеживания, кому уже отправлено приветствие в рамках текущего запуска.
welcomed_users = set()

def get_updates(offset=None):
    url = TELEGRAM_URL + "getUpdates"
    params = {
        "timeout": 100,
        "offset": offset
    }
    response = requests.get(url, params=params)
    return response.json()

def send_message(chat_id, text, reply_markup=None):
    url = TELEGRAM_URL + "sendMessage"
    payload = {
        "chat_id": chat_id,
        "text": text
    }
    if reply_markup:
        payload["reply_markup"] = reply_markup
    requests.post(url, json=payload)

def answer_callback(callback_query_id):
    url = TELEGRAM_URL + "answerCallbackQuery"
    payload = {"callback_query_id": callback_query_id}
    requests.post(url, json=payload)

def delete_message(chat_id, message_id):
    """Удаляет сообщение по заданному chat_id и message_id."""
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
        if response.status_code == 200:
            return {"success": True, "data": response.json()}
        elif response.status_code == 401:
            return {"success": False, "error": "Пользователь не найден"}
        else:
            return {"success": False, "error": f"Неожиданный статус: {response.status_code}"}
    except Exception as e:
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
        response = requests.post(url, json=payload, headers=headers)
        if response.status_code == 200:
            return {"success": True, "data": response.json()}
        else:
            return {"success": False, "error": response.json().get("error", f"Статус: {response.status_code}")}
    except Exception as e:
        return {"success": False, "error": str(e)}

def join_team_request(chat_id, invite_code):
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
        elif response.status_code == 409:
            return {"success": False, "error": "Вы уже присоединились к команде!"}
        elif response.status_code == 404:
            error_data = response.json()
            error_code = error_data.get("code")
            if error_code == "INVITE_CODE_INVALID":
                return {"success": False, "error": "Неверный код приглашения."}
            elif error_code == "USER_NOT_FOUND":
                return {"success": False, "error": "Пользователь не найден. Зарегистрируйтесь через бота."}
        else:
            return {"success": False, "error": f"Ошибка присоединения к команде: Статус: {response.status_code}"}
    except Exception as e:
        return {"success": False, "error": str(e)}

def create_team_request(chat_id, team_name, team_description):
    url = f"{BACKEND_BASE_URL}/team"
    params = {"telegram_id": str(chat_id)}
    payload = {"name": team_name, "description": team_description}
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

def send_welcome(chat_id):
    """Отправляет приветственное сообщение с кнопкой 'Начать'."""
    keyboard = {
        "inline_keyboard": [
            [{"text": "Начать", "callback_data": "start"}]
        ]
    }
    send_message(chat_id, "Добро пожаловать! Нажмите кнопку 'Начать'", reply_markup=keyboard)

def send_main_menu(chat_id):
    """Отправляет главное меню с доступными действиями."""
    keyboard = {
        "inline_keyboard": [
            [{"text": "Присоединиться к команде", "callback_data": "join_team"}],
            [{"text": "Мой профиль", "callback_data": "check_role"}],
        ]
    }
    send_message(chat_id, "Главное меню:", reply_markup=keyboard)

def process_command(command, chat_id):
    if command.startswith('/start'):
        parts = command.split()
        if len(parts) > 1:
            # Если есть параметр для приглашения
            invite_code = parts[1]
            result = join_team_request(chat_id, invite_code)
            if result["success"]:
                send_message(chat_id, "Вы успешно присоединились к команде!")
                send_main_menu(chat_id)
            else:
                send_message(chat_id, result["error"])
        else:
            send_welcome(chat_id)
    elif command == '/help':
        send_message(chat_id, "Доступные команды: /start, /help, /menu")
    elif command == '/menu':
        send_main_menu(chat_id)
    else:
        send_message(chat_id, f"Неизвестная команда: {command}")

def process_user_state(chat_id, text):
    """
    Обрабатывает состояния:
    - awaiting_name – ввод имени при регистрации.
    - awaiting_team_name/awaiting_team_description – создание команды.
    - awaiting_invite_code – присоединение к команде.
    """
    state_info = user_states.get(chat_id)
    if not state_info:
        return False

    state = state_info["state"]

    if state == "awaiting_name":
        # Сохраняем имя и предлагаем выбрать роль кнопками
        user_states[chat_id] = {"state": "awaiting_role", "data": {"name": text}}
        keyboard = {
            "inline_keyboard": [
                [
                    {"text": "Руководитель", "callback_data": "role_manager"},
                    {"text": "Участник", "callback_data": "role_member"}
                ]
            ]
        }
        send_message(chat_id, "Спасибо! Теперь выберите вашу роль:", reply_markup=keyboard)
        return True

    elif state == "awaiting_team_name":
        user_states[chat_id]["data"]["team_name"] = text
        user_states[chat_id]["state"] = "awaiting_team_description"
        send_message(chat_id, "Введите описание команды:")
        return True

    elif state == "awaiting_team_description":
        user_states[chat_id]["data"]["team_description"] = text
        team_name = user_states[chat_id]["data"].get("team_name")
        team_description = text
        result = create_team_request(chat_id, team_name, team_description)
        if result["success"]:
            data = result["data"]
            send_message(
                chat_id,
                f"Команда создана!\nНазвание: {data.get('name')}\nОписание: {data.get('description')}\nСсылка для приглашения: {data.get('invitelink')}"
            )
        else:
            send_message(chat_id, f"Ошибка создания команды: {result['error']}")
        user_states.pop(chat_id, None)
        return True

    elif state == "awaiting_invite_code":
        invite_code = text.strip()
        result = join_team_request(chat_id, invite_code)
        if result["success"]:
            send_message(chat_id, "Вы успешно присоединились к команде!")
        else:
            send_message(chat_id, result["error"])
        user_states.pop(chat_id, None)
        return True

    return False

def process_callback(callback):
    """
    Обрабатывает callback-запросы от inline-кнопок.
    После обработки кнопки удаляет исходное сообщение с кнопкой.
    """
    callback_id = callback["id"]
    chat_id = callback["message"]["chat"]["id"]
    message_id = callback["message"]["message_id"]
    data = callback.get("data")
    answer_callback(callback_id)
    # Удаляем сообщение с кнопкой после нажатия
    delete_message(chat_id, message_id)

    if data == "start":
        # При нажатии кнопки "Начать" проверяем регистрацию
        result = auth_get_request(chat_id)
        if result["success"]:
            send_message(chat_id, "Вы уже зарегистрированы!")
            send_main_menu(chat_id)
        else:
            send_message(chat_id, "Введите, пожалуйста, ваше имя:")
            user_states[chat_id] = {"state": "awaiting_name", "data": {}}
        return

    # Обработка выбора роли при регистрации
    if data.startswith("role_"):
        role = data.split("_", 1)[1]
        state_info = user_states.get(chat_id)
        if state_info and state_info["state"] == "awaiting_role":
            name = state_info["data"].get("name")
            result = auth_post_request(chat_id, name, role)
            if result["success"]:
                send_message(chat_id, "Вы успешно зарегистрированы и авторизованы!")
                # Сохраняем выбранную роль локально
                registered_users[chat_id] = role
                user_states.pop(chat_id, None)
                send_main_menu(chat_id)
            else:
                send_message(chat_id, f"Ошибка: {result['error']}")
        return

    if data == "create_team":
        user_states[chat_id] = {"state": "awaiting_team_name", "data": {}}
        send_message(chat_id, "Введите название команды:")
    elif data == "join_team":
        user_states[chat_id] = {"state": "awaiting_invite_code", "data": {}}
        send_message(chat_id, "Ожидаю ссылку для присоединения команды:")
    elif data == "check_role":
        # Сначала пытаемся получить роль из локального словаря
        role = registered_users.get(chat_id)
        if not role:
            auth_data = auth_get_request(chat_id)
            if auth_data["success"]:
                role = auth_data["data"].get("role")
        role_map = {"manager": "Руководитель", "member": "Участник"}
        display_role = role_map.get(role, role if role else "неизвестно")
        send_message(chat_id, f"Ваша роль: {display_role}")
        send_main_menu(chat_id)

def main():
    offset = None
    print("Бот запущен и ожидает обновлений...")

    # Если указан ADMIN_CHAT_ID, отправляем ему приветствие сразу при запуске
    if ADMIN_CHAT_ID:
        try:
            chat_id = int(ADMIN_CHAT_ID)
            send_welcome(chat_id)
            welcomed_users.add(chat_id)
        except Exception as e:
            print("Ошибка отправки приветствия:", e)

    while True:
        updates = get_updates(offset)
        print("Получены обновления:", updates)  # для отладки
        if updates.get("ok"):
            for update in updates["result"]:
                offset = update["update_id"] + 1

                if "callback_query" in update:
                    process_callback(update["callback_query"])
                    continue

                message = update.get("message")
                if not message:
                    continue

                chat_id = message["chat"]["id"]
                text = message.get("text")
                if not text:
                    continue

                print(f"Новое сообщение от {chat_id}: {text}")

                # Если пользователь ранее не был приветствован, отправляем ему сообщение с кнопкой "Начать"
                if chat_id not in welcomed_users:
                    send_welcome(chat_id)
                    welcomed_users.add(chat_id)
                    continue

                if chat_id in user_states:
                    if process_user_state(chat_id, text):
                        continue

                if text.startswith('/'):
                    process_command(text, chat_id)
                else:
                    send_message(chat_id, "Пожалуйста, используйте команды, начинающиеся с '/' или кнопки для навигации.")
        time.sleep(1)

if __name__ == '__main__':
    main()
