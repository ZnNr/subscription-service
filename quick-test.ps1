# quick-test.ps1
Write-Host "=== БЫСТРЫЙ ТЕСТ API ===" -ForegroundColor Green

# 1. Создание подписки
$body = @{
    service_name = "Test Service"
    price = 100
    user_id = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
    start_date = "01-2025"
} | ConvertTo-Json

Write-Host "1. Создание подписки..." -ForegroundColor Cyan
try {
    $result = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/subscriptions" `
        -Method Post `
        -Body $body `
        -ContentType "application/json"

    Write-Host "   ✅ Успешно! ID: $($result.id)" -ForegroundColor Green
    $testId = $result.id
} catch {
    Write-Host "   ❌ Ошибка: $($_.Exception.Message)" -ForegroundColor Red
    exit
}

# 2. Получение всех подписок
Write-Host "`n2. Получение всех подписок..." -ForegroundColor Cyan
try {
    $all = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/subscriptions" -Method Get
    Write-Host "   ✅ Найдено: $($all.Count) подписок" -ForegroundColor Green
} catch {
    Write-Host "   ❌ Ошибка: $($_.Exception.Message)" -ForegroundColor Red
}

# 3. Подсчет суммы
Write-Host "`n3. Подсчет суммы..." -ForegroundColor Cyan
$summaryBody = @{
    start_date = "01-2025"
    end_date = "12-2025"
} | ConvertTo-Json

try {
    $summary = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/subscriptions/summary" `
        -Method Post `
        -Body $summaryBody `
        -ContentType "application/json"

    Write-Host "   ✅ Сумма: $($summary.total_amount) руб." -ForegroundColor Green
    Write-Host "   ✅ Количество: $($summary.count)" -ForegroundColor Green
} catch {
    Write-Host "   ❌ Ошибка: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "`n=== ТЕСТ ЗАВЕРШЕН ===" -ForegroundColor Green