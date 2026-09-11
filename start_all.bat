@echo off
rem ============================================================
rem  B2B Marketing Agent Harness - one-click start
rem  - Backend : go run ./cmd/api  -> http://localhost:8090
rem  - Frontend: cd web && npm run dev -> http://localhost:5173
rem  - No PostgreSQL / Redis needed out-of-the-box (mock LLM + in-memory store).
rem    To switch to real PG, set BH_DB_DRIVER=postgres in .env first.
rem ============================================================
setlocal

set "ROOT=%~dp0"
set "WEB_DIR=%ROOT%web"
set "API_ADDR=:8090"
set "WEB_PORT=5173"

echo ============================================================
echo   B2B Marketing Agent Harness - One-Click Start
echo ------------------------------------------------------------
echo   Backend API : http://localhost%API_ADDR%/api/v1/health
echo   Workbench UI: http://localhost:%WEB_PORT%
echo ============================================================
echo.

where go >nul 2>nul
if errorlevel 1 (
  echo [ERROR] go not found. Install Go 1.23+ and add to PATH.
  pause
  exit /b 1
)
where node >nul 2>nul
if errorlevel 1 (
  echo [ERROR] node not found. Install Node 18+ and add to PATH.
  pause
  exit /b 1
)
if not exist "%WEB_DIR%\package.json" (
  echo [ERROR] frontend project not found: %WEB_DIR%\package.json
  pause
  exit /b 1
)

if not exist "%WEB_DIR%\node_modules" (
  echo [setup] First run detected: installing frontend dependencies ...
  pushd "%WEB_DIR%"
  call npm install --no-audit --no-fund
  popd
  if errorlevel 1 (
    echo [ERROR] npm install failed.
    pause
    exit /b 1
  )
)

echo [start] Starting backend API in a new window ...
cd /d "%ROOT%"
start "harness-api" cmd /k "title harness-api && echo http://localhost%API_ADDR%/api/v1/health && go run ./cmd/api -config configs/config.yaml"

echo [start] Starting Workbench UI in a new window ...
cd /d "%WEB_DIR%"
start "harness-workbench" cmd /k "title harness-workbench && echo http://localhost:%WEB_PORT% && npm run dev"

echo.
echo [OK] Both services launched in new windows. Close those windows to stop them.
echo.
echo Tips:
echo   - Backend runs in mock mode by default (no PG / Redis / LLM keys needed).
echo   - To switch to real LLM: set BH_LLM_PROVIDER=openai + BH_LLM_API_KEY=... in .env
echo   - To switch to real PostgreSQL: set BH_DB_DRIVER=postgres in .env
endlocal
