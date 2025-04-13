# 🧩 micro-backend

以 Go + Gin 架構擴延的微型後端專案，使用 MySQL 資料庫與 JWT 作為驗證方式。支援帳號註冊與登入功能，並透過 Docker Compose 快速建立完整開發環境。

---

## 📦 專案結構

```
micro-backend/
├── Dockerfile
├── docker-compose.yml
├── .env
├── .env.example         ← ✅ 範例環境變數檔（不含敏感資訊）
├── .gitignore           ← ✅ 應包含 `.env`
├── .air.toml
├── main.go
├── models/
│   └── user.go
├── handlers/
│   └── auth.go
├── middlewares/
│   └── jwt.go           ← ✅ JWT 驗證中介層
├── migrations/
│   ├── 000001_create_users.up.sql
│   └── 000001_create_users.down.sql
```

---

## 🚀 專案啟動方式（開發）

### 1️⃣ 建立 `.env` 檔案於根目錄

請自行建立 `.env` 檔，根據下面格式填入你的環境值：

```env
# 資料庫設定
MYSQL_ROOT_PASSWORD=your_root_password
MYSQL_DATABASE=your_database
MYSQL_USER=your_db_user
MYSQL_PASSWORD=your_db_password

# Go 應用環境變數
DB_HOST=db
DB_PORT=3306
DB_USER=your_db_user
DB_PASSWORD=your_db_password
DB_NAME=your_database
PORT=8088
JWT_SECRET=your_jwt_secret_key
```

> ✉️ 建議使用 `.env.example` 作為格式模板，並在 `.gitignore` 中掛上 `.env`，避免故意上傳到 GitHub

---

### 2️⃣ 啟動 MySQL + Go 應用（含 hot reload）

```bash
docker compose up --build
```

後端會自動使用 air 啟動，支援即時重編譯。

---

### 3️⃣ 初始化資料表（僅首次）

```bash
migrate -path ./migrations -database "mysql://user:pass@tcp(localhost:3306)/app_db" up
```

> 若重置資料庫，可加 `down -v` 並重新 migrate。

---

## 🔐 JWT Middleware

已實作 JWT 驗證中介層，使用者登入取得 Token 後，需透過 `Authorization: Bearer <token>` 才能存取受保護的路由。

範例受保護路由：

### 🔒 取得使用者個人資訊 `/api/v1/profile`
```bash
curl -X GET http://localhost:8088/api/v1/profile \
  -H "Authorization: Bearer <your_token>"
```

成功回應：
```json
{
  "user_id": 1,
  "email": "w@w.com",
  "username": "walter"
}
```

---

## 🧪 API 測試指令

### ➕ 註冊帳號
```bash
curl -X POST http://localhost:8088/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{"username":"walter","email":"w@w.com","password":"123456"}'
```

### 🔐 登入帳號（取得 JWT）
```bash
curl -X POST http://localhost:8088/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"email":"w@w.com","password":"123456"}'
```

成功回傳：
```json
{ "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6..." }
```

### 🧍 取得個人資訊（JWT 驗證）
```bash
curl -X GET http://localhost:8088/api/v1/profile \
  -H "Authorization: Bearer <token>"
```

---

## 🛠 開發常用指令

| 指令                            | 說明                      |
|------------------------------------|-------------------------------|
| `docker compose up --build`        | 啟動後端與資料庫          |
| `docker compose down -v`           | 關閉並刪除 volume（清資料） |
| `docker logs go-app`               | 查看後端 log               |
| `docker image prune -f`            | 清除無用 image            |

---

## 🧰 開發用資料庫連線指令

進入 MySQL container：
```bash
docker exec -it mysql mysql -u user -p
# 密碼：<your_db_password>
```

常用指令：
```sql
USE app_db;
SHOW TABLES;
SELECT * FROM users;
```

---

## ✅ 注意事項

- MySQL 採用 volume 保留資料，除非使用 `-v` 清除，否則不需重跑 migration
- Go 程式內建 DB 重試機制，避免因 container 尚未啟動導致連線錯誤
- 所有環境變數集中在 `.env`，不應確認內容放入他人的 Git repo

---

📬 若你需要加入更多 API、JWT 權限群組或自動 migration 機制，可參考進階章節或擴充分支。

