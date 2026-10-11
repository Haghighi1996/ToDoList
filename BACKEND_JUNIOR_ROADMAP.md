# نقشهٔ راه تکمیل Todo API

این سند برای تبدیل وضعیت فعلی پروژه به یک نمونه‌کار بک‌اند در سطح **Junior Backend Developer** نوشته شده است. هدف، ساخت یک REST API کوچک اما تمیز، قابل اجرا، تست‌پذیر و قابل ارائه است؛ نه ساختن یک سیستم پیچیدهٔ سازمانی.

## 1. وضعیت فعلی پروژه

پروژه با Go، PostgreSQL و `golang-migrate` شروع شده و لایه‌های زیر را دارد:

- `model`: مدل‌های `User`، `Session` و `Todo`
- `repository`: ارتباط SQL با دیتابیس
- `service`: بخشی از منطق Todo و Auth
- `middleware`: middleware احراز هویت، recovery و request ID
- `migrations`: ایجاد جدول‌ها
- `cmd/migrate`: اجرای migrationها

این لایه‌بندی، برای یک پروژهٔ جونیور انتخاب خوبی است. اما API هنوز قابل استفاده نیست، چون entrypoint سرور، router و handlerها وجود ندارند و چند باگ مهم در auth و migration وجود دارد.

## 2. مشکلات فعلی و اولویت حل

### P0 — قبل از هر قابلیت جدید

| مشکل | اثر | راه‌حل پیشنهادی |
| --- | --- | --- |
| توکن خام در middleware جست‌وجو می‌شود ولی در دیتابیس hash ذخیره شده است | تمام endpointهای محافظت‌شده `401` می‌شوند | تابع `HashToken` را در یک پکیج مشترک auth قرار بده و قبل از `FindActiveByTokenHash` آن را صدا بزن |
| نبودن user با `sql.ErrNoRows` برمی‌گردد، ولی سرویس دنبال `repository.ErrNotFound` می‌گردد | ثبت‌نام/ورود اولین دستگاه شکست می‌خورد | یک `repository.ErrNotFound` مشترک تعریف و همهٔ repositoryها `sql.ErrNoRows` را به آن تبدیل کنند |
| نام فایل migrationها با قرارداد `golang-migrate` سازگار نیست | migration ممکن است اصلاً شناسایی نشود | فایل‌ها را به `001_create_users.up.sql` و فایل‌های down متناظر تغییر نام بده |
| `gen_random_uuid()` بدون فعال‌کردن `pgcrypto` استفاده شده است | migration دیتابیس خالی خطا می‌دهد | در اولین migration `CREATE EXTENSION IF NOT EXISTS pgcrypto;` اضافه کن |
| سرور، router و HTTP handler وجود ندارد | کل پروژه endpoint قابل فراخوانی ندارد | `cmd/api/main.go`، router و handlerها را ایجاد کن |

### P1 — برای کیفیت API

| مشکل | راه‌حل پیشنهادی |
| --- | --- |
| پاک‌کردن `due_date` در update ممکن نیست | برای DTO یک type اختیاری بساز که تفاوت «ارسال نشده»، `null` و مقدار واقعی را نگه دارد؛ یا endpoint جدا برای clear کردن تاریخ داشته باش |
| متن فقط شامل فاصله در update خطای دیتابیس می‌دهد | همیشه قبل از validation از `strings.TrimSpace` استفاده کن و اگر خالی بود `400 invalid_request` بده |
| `Delete` همهٔ خطاها را `not_found` می‌کند | فقط `repository.ErrNotFound` را به 404 تبدیل کن؛ خطاهای دیگر را log کن و 500 بده |
| جست‌وجوی session ایندکس ندارد | روی `sessions(token_hash)` unique index ایجاد کن |
| `updated_at` فقط در query تغییر می‌کند | برای این پروژه کافی است؛ اگر منطق زیاد شد trigger PostgreSQL یا یک روش واحد برای update تعریف کن |
| رمز دیتابیس در کد پیش‌فرض است | `.env.example` بساز؛ در production مقدارهای لازم را اجباری کن و secret را commit نکن |

### P2 — برای ارائهٔ حرفه‌ای‌تر

- README با راه‌اندازی، متغیرهای محیطی و نمونهٔ API.
- تست unit برای serviceها و تست integration برای repositoryها.
- GitHub Actions برای `go test ./...` و `go vet ./...`.
- logging ساخت‌یافته با request ID.
- pagination، فیلتر completed/overdue و sort مستندشده.
- health check برای بررسی اتصال اپلیکیشن و دیتابیس.

## 3. قرارداد API پیشنهادی

همهٔ پاسخ‌ها JSON هستند و خطاها ساختار ثابت دارند:

```json
{
  "error": "invalid_request",
  "message": "text is required"
}
```

| Method | Path | Auth | کاربرد |
| --- | --- | --- | --- |
| `POST` | `/api/v1/auth/anonymous` | خیر | ساخت یا ورود کاربر ناشناس با `device_id` |
| `POST` | `/api/v1/auth/logout` | بله | باطل‌کردن session فعلی |
| `GET` | `/api/v1/todos` | بله | فهرست Todoها با pagination و filter |
| `POST` | `/api/v1/todos` | بله | ساخت Todo |
| `GET` | `/api/v1/todos/{id}` | بله | دریافت یک Todo |
| `PATCH` | `/api/v1/todos/{id}` | بله | ویرایش Todo |
| `DELETE` | `/api/v1/todos/{id}` | بله | حذف Todo |
| `GET` | `/healthz` | خیر | سلامت سرویس |

نمونهٔ ایجاد Todo:

```http
POST /api/v1/todos
Authorization: Bearer <token>
Content-Type: application/json

{"text":"یادگیری Go","due_date":"2026-10-15T12:00:00Z"}
```

فیلتر لیست:

```text
GET /api/v1/todos?completed=false&overdue=true&sort=due_date_asc&page=1&limit=20
```

قواعد پیشنهادی:

- `text` بعد از trim باید بین 1 تا 500 کاراکتر باشد.
- `page` حداقل 1 و `limit` بین 1 تا 100 باشد.
- کاربر فقط Todoهای خودش را می‌بیند یا تغییر می‌دهد.
- token تصادفی 32-byte باشد، فقط hash آن ذخیره شود و 24 ساعت اعتبار داشته باشد.

## 4. ساختار پیشنهادی پروژه

```text
todo-api/
├── cmd/
│   ├── api/
│   │   └── main.go                 # ساخت dependencyها و اجرای HTTP server
│   └── migrate/
│       └── main.go                 # اجرای migrationها
├── internal/
│   ├── config/
│   │   └── config.go               # خواندن و اعتبارسنجی env
│   ├── db/
│   │   └── postgres.go             # ساخت و بستن connection pool
│   ├── model/
│   │   ├── user.go
│   │   ├── session.go
│   │   └── todo.go
│   ├── dto/
│   │   ├── auth.go                 # request/response HTTP
│   │   ├── todo.go
│   │   └── error.go
│   ├── repository/
│   │   ├── errors.go               # ErrNotFound مشترک
│   │   ├── interfaces.go
│   │   ├── postgres_user.go
│   │   ├── postgres_session.go
│   │   └── postgres_todo.go
│   ├── service/
│   │   ├── errors.go                # AppError و mapping خطاها
│   │   ├── auth_service.go
│   │   └── todo_service.go
│   ├── handler/
│   │   ├── auth_handler.go          # decode/validate/call service/encode
│   │   ├── todo_handler.go
│   │   ├── health_handler.go
│   │   └── response.go              # JSON و پاسخ خطا یکپارچه
│   ├── middleware/
│   │   ├── auth.go
│   │   ├── request_id.go
│   │   ├── logging.go
│   │   └── recovery.go
│   ├── router/
│   │   └── router.go                # ثبت routeها و middlewareها
│   └── auth/
│       └── token.go                 # GenerateToken و HashToken مشترک
├── migrations/
│   ├── 001_enable_pgcrypto.up.sql
│   ├── 001_enable_pgcrypto.down.sql
│   ├── 002_create_users.up.sql
│   ├── 002_create_users.down.sql
│   ├── 003_create_sessions.up.sql
│   ├── 003_create_sessions.down.sql
│   ├── 004_create_todos.up.sql
│   └── 004_create_todos.down.sql
├── tests/
│   └── integration/                 # در صورت نیاز به تست end-to-end
├── docker-compose.yml
├── Dockerfile
├── .env.example
├── .gitignore
├── Makefile
├── README.md
└── go.mod
```

> لازم نیست همه‌چیز را یک‌جا بسازی. برای سطح جونیور، مهم‌تر از تعداد پوشه‌ها این است که مسئولیت هر لایه روشن و تکرار کد کم باشد.

## 5. مسئولیت هر لایه

```text
HTTP Request
    ↓
Middleware (request ID, recovery, auth)
    ↓
Handler (decode JSON, parse params, ارسال status code)
    ↓
Service (قانون کسب‌وکار و validation)
    ↓
Repository (SQL و تبدیل نتیجه به model)
    ↓
PostgreSQL
```

- **Handler** نباید SQL بنویسد یا تصمیم کسب‌وکاری پیچیده بگیرد.
- **Service** نباید `http.ResponseWriter` بشناسد.
- **Repository** نباید status code برگرداند.
- **Model** ساختار دادهٔ دامنه است؛ **DTO** قرارداد ورودی/خروجی HTTP است.

## 6. نحوهٔ پیاده‌سازی قدم‌به‌قدم

### مرحله 1: قابل اجرا کردن دیتابیس و migration

1. نام migrationها را به قرارداد `.up.sql` و `.down.sql` تغییر بده.
2. migration افزونهٔ `pgcrypto` را قبل از ساخت جدول users قرار بده.
3. برای sessionها index زیر را اضافه کن:

```sql
CREATE UNIQUE INDEX idx_sessions_token_hash ON sessions(token_hash);
```

4. در `.env.example` این متغیرها را مستند کن:

```env
PORT=8080
DATABASE_URL=postgres://todo_user:todo_password@localhost:5432/todo_api?sslmode=disable
```

5. migration را روی یک volume کاملاً تازه اجرا کن؛ فقط اجرای migration روی دیتابیس قدیمی کافی نیست.

### مرحله 2: درست‌کردن Auth

1. در `internal/auth/token.go` دو تابع `GenerateToken()` و `HashToken(token string)` قرار بده.
2. هم AuthService و هم Auth middleware فقط از همین تابع Hash استفاده کنند.
3. در هر repository، `sql.ErrNoRows` را به `repository.ErrNotFound` تبدیل کن.
4. در middleware، header را دقیق بررسی کن: فقط فرم `Bearer <token>` معتبر باشد.
5. handler ورود، `device_id` خالی یا بیشتر از 64 کاراکتر را با 400 رد کند.
6. برای logout، session پیدا شده بر اساس hash توکن را revoke کن.

### مرحله 3: ساخت HTTP server و router

در `cmd/api/main.go` به‌ترتیب زیر عمل کن:

1. config را بارگذاری و اعتبارسنجی کن.
2. اتصال دیتابیس را باز کن و با `defer Close` ببند.
3. repositoryها را بساز.
4. serviceها را با repositoryهای مورد نیاز بساز.
5. handlerها و router را بساز.
6. سرور HTTP با timeoutهای read/write/idle اجرا کن.
7. با `signal.NotifyContext` خاموش‌شدن graceful را پیاده کن.

برای router می‌توانی از `net/http` و `ServeMux` استفاده کنی. برای این پروژه نصب framework جدید ضروری نیست. اگر از `chi` استفاده می‌کنی، دلیلش را در README بنویس و همان انتخاب را تا انتها ثابت نگه دار.

### مرحله 4: Handlerها و پاسخ استاندارد

یک helper مانند `writeJSON(w, status, data)` و `writeError(w, appError)` بساز. هر handler باید این الگو را دنبال کند:

1. body را با محدودیت اندازه decode کن.
2. path/query parameter را parse کن.
3. service را صدا بزن.
4. خطای domain را به status code مناسب تبدیل کن.
5. پاسخ JSON بده.

Status codeهای کافی برای این پروژه:

- `201 Created`: ساخت Todo یا session
- `200 OK`: خواندن/ویرایش/logout
- `204 No Content`: حذف
- `400 Bad Request`: ورودی نامعتبر
- `401 Unauthorized`: token نامعتبر یا منقضی
- `404 Not Found`: Todo وجود ندارد یا متعلق به کاربر نیست
- `500 Internal Server Error`: خطای غیرمنتظره؛ جزئیات آن فقط log شود

### مرحله 5: Update صحیح Todo

برای PATCH، DTO سادهٔ pointerدار نمی‌تواند `null` و فیلد غایب را از هم جدا کند. یکی از این دو راه را انتخاب کن:

1. یک `OptionalTime` با `UnmarshalJSON` بنویس که `Set` و `Value` داشته باشد؛ `Set=true, Value=nil` یعنی پاک‌کردن تاریخ.
2. برای پروژهٔ ساده‌تر، `due_date` را فقط با endpoint جداگانه پاک کن، مثلاً `DELETE /api/v1/todos/{id}/due-date`.

راه اول برای تمرین Go ارزشمندتر است، اما باید تست شود.

### مرحله 6: تست

حداقل تست‌های زیر را اضافه کن:

| نوع | سناریو |
| --- | --- |
| Unit: TodoService | متن خالی، فقط فاصله، بیش از 500 کاراکتر، ساخت موفق |
| Unit: TodoService | تغییر completed، پاک‌کردن due date، Todo پیدا نشد |
| Unit: AuthService | کاربر جدید، کاربر موجود، token تصادفی، logout |
| Unit: middleware | header غایب، فرمت بد Bearer، token هش‌شدهٔ معتبر |
| Integration: repository | ایجاد، لیست با فیلتر، محدودبودن داده به user خودش |
| HTTP | مسیر موفق و پاسخ‌های 400/401/404 |

اول repositoryها را با interface mock نکن؛ serviceها را با fake repository ساده تست کن. برای integration test یک دیتابیس جدا با `TEST_DATABASE_URL` داشته باش و هرگز روی دیتابیس توسعهٔ اصلی تست destructive اجرا نکن.

## 7. ابزارهای کمکی پیشنهادی

### Makefile

دستورهای زیر کافی‌اند:

```text
make run       # اجرای API
make migrate-up
make migrate-down
make test
make vet
make fmt
make docker-up
```

### .gitignore

حداقل این موارد را نادیده بگیر:

```gitignore
.env
bin/
coverage.out
migrate
```

### README

باید شامل این بخش‌ها باشد:

1. مسئله و قابلیت‌ها
2. پیش‌نیازها: Go و Docker
3. راه‌اندازی قدم‌به‌قدم
4. متغیرهای محیطی
5. اجرای migration
6. endpointها با نمونهٔ درخواست
7. اجرای testها
8. محدودیت‌های فعلی و ایده‌های توسعه

## 8. تعریف «تمام‌شده» برای نسخهٔ اول

نسخهٔ اول وقتی تمام‌شده محسوب می‌شود که:

- [ ] دیتابیس تازه با یک دستور بالا بیاید و migration با موفقیت اجرا شود.
- [ ] کاربر ناشناس token معتبر دریافت کند.
- [ ] token نامعتبر یا منقضی 401 برگرداند.
- [ ] همهٔ CRUDهای Todo کار کنند و دسترسی بین کاربران ممکن نباشد.
- [ ] pagination، فیلتر completed و overdue، و sort کار کنند.
- [ ] ورودی نامعتبر 400 و Todo ناموجود 404 برگرداند.
- [ ] تست‌ها و `go vet ./...` پاس شوند.
- [ ] README یک نفر دیگر را بدون کمک تو بتواند پروژه را اجرا کند.

## 9. ترتیب پیشنهادی کار

1. اصلاح migration و اجرای آن روی دیتابیس تازه.
2. رفع دو باگ Auth و نوشتن تست آن‌ها.
3. اضافه‌کردن server، router و health check.
4. ساخت handlerهای auth و Todo.
5. تکمیل PATCH و error mapping.
6. تست‌های service و HTTP.
7. README، Makefile، `.env.example` و CI.

اگر این ترتیب را رعایت کنی، هر مرحله یک خروجی قابل اجرا دارد و در انتها پروژه‌ای داری که هم در مصاحبه قابل توضیح است و هم نشان می‌دهد اصول پایهٔ بک‌اند—لایه‌بندی، API design، امنیت session، migration و test—را می‌شناسی.
