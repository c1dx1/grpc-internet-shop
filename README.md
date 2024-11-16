# Интернет-магазин (Microservices Architecture)

Проект представляет собой интернет-магазин, построенный на основе микросервисной архитектуры. Система включает несколько независимых сервисов, каждый из которых отвечает за свою часть функционала. Основные технологии: **Go**, **gRPC**, **RabbitMQ**, **PostgreSQL**, **Redis**, **Docker**.

## Архитектура

Проект состоит из следующих микросервисов:

1. **User Service**: Регистрация, аутентификация пользователей, управление сессиями (Redis для сессий).
2. **Product Service**: Управление товарами магазина (CRUD операции для товаров).
3. **Cart Service**: Управление корзиной покупок пользователя.
4. **Order Service**: Оформление и отслеживание заказов.
5. **Notification Service**(в разработке): Отправка уведомлений о действиях пользователя (например, заказ оформлен).
6. **Gateway**: Веб-интерфейс, служащий точкой входа для всех клиентских запросов (через gRPC).

## Используемые технологии

- **Go**: Основной язык для написания микросервисов.
- **gRPC**: Взаимодействие между сервисами.
- **RabbitMQ**: Для обмена событиями между микросервисами (используется в основном для уведомлений).
- **PostgreSQL**: Основная база данных для хранения информации о продуктах, заказах, корзинах, уведомлений и пользователях.
- **Redis**: Хранение сессий пользователей.
- **Docker**: Контейнеризация всех микросервисов.

## Функциональность

### 1. User Service
- Регистрация пользователей.
- Аутентификация и управление сессиями (с использованием Redis).
- Выход из аккаунта (удаление сессии).

### 2. Product Service
- Получение списка товаров.
- Получение информации о конкретном товаре.
- Управление товарами (для администратора).

### 3. Cart Service
- Добавление товаров в корзину.
- Удаление товаров из корзины.
- Просмотр текущей корзины.

### 4. Order Service
- Оформление заказа.
- Получение информации о заказе по ID.

### 5. Notification Service
- Получение событий из других сервисов через RabbitMQ (например, оформление заказа).
- Отправка уведомлений пользователям по email.

### 6. Gateway Service
- Веб-интерфейс, доступ к которому предоставляется через REST API (с использованием Gin).
- Аутентификация пользователя через middleware.
- Взаимодействие с микросервисами через gRPC.

