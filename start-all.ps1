# Start all three services in parallel

Write-Host "Starting Go Backend (port 8080)..." -ForegroundColor Green
Start-Process powershell -ArgumentList "cd cmd/api; go run main.go"

Write-Host "Starting API Gateway (port 3000)..." -ForegroundColor Cyan
Start-Process powershell -ArgumentList "cd api-gateway; npm run dev"

Write-Host "Starting Frontend (port 3001)..." -ForegroundColor Yellow
Start-Process powershell -ArgumentList "cd frontend; npm run dev"

Write-Host "All services starting..." -ForegroundColor Magenta
Write-Host "Go Backend:  http://localhost:8080" -ForegroundColor Green
Write-Host "API Gateway: http://localhost:3000" -ForegroundColor Cyan
Write-Host "Frontend:    http://localhost:3001" -ForegroundColor Yellow
