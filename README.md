# Todo API

یک REST API ساده برای مدیریت کارهای روزانه، نوشته‌شده با Go و PostgreSQL. این پروژه از session-based token authentication استفاده می‌کند: توکن تصادفی به کاربر داده می‌شود و فقط hash آن در دیتابیس ذخیره می‌شود.

## قابلیت‌ها

- ورود یا ساخت کاربر ناشناس با `device_id`
- logout و باطل‌کردن session
- ساخت، مشاهده، فهرست، ویرایش و حذف Todo
- محدودبودن هر Todo به کاربر صاحب آن
- pagination، فیلتر `completed` و `overdue`، و مرتب‌سازی تاریخ سررسید
- migration دیتابیس، health check و graceful shutdown

## پیش‌نیازها

- Go 1.26 یا جدیدتر
- Docker و Docker Compose (برای اجرای PostgreSQL)

## راه‌اندازی محلی

```bash
cp .env.example .env
docker compose up -d postgres
export DATABASE_URL='postgres://todo_user:todo_password@localhost:5432/todo_api?sslmode=disable'
make migrate-up
make run
```

سرویس در `http://localhost:8080` اجرا می‌شود. برای اجرای کامل با Docker:

```bash
docker compose up --build
```

## متغیرهای محیطی

| نام | اجباری | نمونه | توضیح |
| --- | --- | --- | --- |
| `PORT` | خیر | `8080` | پورت HTTP |
| `DATABASE_URL` | بله | `postgres://...` | connection string PostgreSQL |

## API

### دریافت token

```bash
curl -X POST http://localhost:8080/api/v1/auth/anonymous \
  -H 'Content-Type: application/json' \
  -d '{"device_id":"my-phone-001"}'
```

پاسخ شامل `token` است. آن را در درخواست‌های بعدی به شکل زیر بفرستید:

```text
Authorization: Bearer <token>
```

### Todoها

| Method | Path | توضیح |
| --- | --- | --- |
| `GET` | `/healthz` | سلامت سرویس و دیتابیس |
| `POST` | `/api/v1/auth/anonymous` | ورود ناشناس |
| `POST` | `/api/v1/auth/logout` | logout |
| `GET` | `/api/v1/todos` | لیست Todoها |
| `POST` | `/api/v1/todos` | ساخت Todo |
| `GET` | `/api/v1/todos/{id}` | یک Todo |
| `PATCH` | `/api/v1/todos/{id}` | ویرایش Todo |
| `DELETE` | `/api/v1/todos/{id}` | حذف Todo |

نمونهٔ ساخت Todo:

```bash
curl -X POST http://localhost:8080/api/v1/todos \
  -H 'Authorization: Bearer <token>' \
  -H 'Content-Type: application/json' \
  -d '{"text":"Finish API README","due_date":"2026-10-15T12:00:00Z"}'
```

برای پاک‌کردن تاریخ سررسید، مقدار `null` ارسال کنید:

```json
{"due_date": null}
```

نمونهٔ فیلتر لیست:

```text
GET /api/v1/todos?completed=false&overdue=true&sort=due_date_asc&page=1&limit=20
```

## دستورهای توسعه

```bash
make fmt
make test
make vet
make migrate-up
make migrate-down
```

## پاسخ خطا

همهٔ خطاهای شناخته‌شده ساختار یکسان دارند:

```json
{"error":"invalid_request","message":"Invalid request"}
```

جزئیات خطاهای غیرمنتظره فقط در log ذخیره می‌شوند و به کاربر نمایش داده نمی‌شوند.
