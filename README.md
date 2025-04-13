# 🤩 micro-backend

以 Go + Gin 構築擴延的微型後端專案，使用 MySQL 資料庫與 JWT 作為驗證方式。支援帳號註冊與登入功能，並通過 Docker Compose 快速建立完整開發環境。

---

## 📦 專案結構

```
micro-backend/
├— Dockerfile
├— docker-compose.yml
├— .env
├— .env.example         ← ✅ 範例環境變數檔（不含敏感資訊）
├— .gitignore           ← ✅ 應包含 `.env`
├— .air.toml
├— main.go
├— docs/                ← ✅ Swagger 文件產出目錄
├— models/
│   └— user.go
├— handlers/
│   ├— auth.go
│   └— profile.go       ← ✅ 使用者資訊 API
├— middlewares/
│   └— jwt.go           ← ✅ JWT 驗證中介層
└— migrations/
    ├— 000001_create_users.up.sql
    └— 000001_create_users.down.sql
```

---

## 🚀 專案啟動方式（開發）

### 1️⃣ 建立 `.env` 檔案於根目錄

請自行建立 `.env` 檔，根據下面格式填入你的環境值：

```env
# 資料庫設定
MYSQL_ROOT_PASSWORD=your_root_password
MYSQL_DATABASE=app_db
MYSQL_USER=your_db_user
MYSQL_PASSWORD=your_db_password

# Go 應用環境變數
DB_HOST=db
DB_PORT=3306
DB_USER=your_db_user
DB_PASSWORD=your_db_password
DB_NAME=app_db
PORT=8088
JWT_SECRET=your_jwt_secret_key
```

> ✉️ 建議使用 `.env.example` 作為格式模板，並在 `.gitignore` 中掛上 `.env`，避免故意上傳到 GitHub

---

### 2️⃣ 啟動 MySQL + Go 應用（含 hot reload）+ 自動執行 migrate

```bash
docker compose up --build
```

- ✅ 使用 Air 自動熱重載
- ✅ 會啟動 `migrate` container，自動執行未跑過的 migration 檔案
- ✅ 保留 DB 資料（volume 機制）

---

### 🔁 安全重啟方式（保留資料）
```bash
docker compose down && docker compose up --build
```

### ⚠️ 若你想重置資料庫（dev 限用）
```bash
docker compose down -v && docker compose up --build
```

> ⚠️ `-v` 會清除 volume（包括 MySQL 資料），僅限測試用

---

## 📘 整合 Swagger 文件（API 規格說明）

### ✅ 安裝工具（僅需一次）
```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

### ✅ 安裝套件
```bash
go get github.com/swaggo/gin-swagger

go get github.com/swaggo/files
```

### ✅ 產生文件（每次更新註解後執行）
```bash
swag init
```

### ✅ 啟動後瀏覽 Swagger UI
```http
http://localhost:8088/swagger/index.html
```

---

## 🔐 JWT Middleware

已實作 JWT 驗證中介層，使用者登入取得 Token 後，需通過 `Authorization: Bearer <token>` 才能存取受保護的路由。

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

## 💪 API 測試指令

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

### 🠍 取得個人資訊（JWT 驗證）
```bash
curl -X GET http://localhost:8088/api/v1/profile \
  -H "Authorization: Bearer <token>"
```

---

## 🛠 開發常用指令

| 指令                            | 說明                          |
|--------------------------------|-------------------------------|
| `docker compose up --build`   | 啟動後端與資料庫 + migrate    |
| `docker compose down`         | 關閉所有服務（保留資料）       |
| `docker compose down -v`      | 關閉並刪除 volume（清資料）   |
| `docker logs go-app`          | 查看後端 log                  |
| `docker image prune -f`       | 清除無用 image                |
| `swag init`                   | 產生 Swagger 文件（docs/）    |

---

## 📃 正式環境指令（Production）

| 指令                                                         | 說明                                |
|--------------------------------------------------------------------|-------------------------------------|
| `docker compose -f docker-compose.prod.yml build --no-cache`      | 建立 production 版本映像（重新編譯） |
| `docker compose -f docker-compose.prod.yml up -d`                 | 背景啟動正式服務                    |
| `docker compose -f docker-compose.prod.yml down`                  | 停止正式服務（保留資料）           |
| `docker compose -f docker-compose.prod.yml down -v`               | 停止並刪除資料 volume（重建資料）  |
| `docker logs go-app`                                              | 查看正式服務 Log                    |
| `curl http://<your_server_ip>:8088/swagger/index.html`            | 確認 Swagger 是否部署成功           |

> ✅ `-f` 是指定用 `docker-compose.prod.yml`，用來與 dev 隔離  
> ✅ `-d` 代表 background mode，不會卡在端末機

---

## 🛠 開發用 MySQL CLI 連線

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
- `migrate` service 會自動讀取 `.env` 並在啟動時執行 `up`
- 本機也應遵守 migrate 遞增原則，請用 `migrate create` 新增版本

---

📬 若你需要加入更多 API、JWT 權限群組或自動 migration 機制，可參考進階章節或擴充分支。

