# Smoke-тест API: прогоняет эталонный расчёт из плана MVP.
# Использование: .\scripts\smoke.ps1 [baseUrl]  (по умолчанию http://localhost:8080)
param([string]$Base = "http://localhost:8080")

[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$ErrorActionPreference = "Stop"
$failed = 0

function Assert($cond, $msg) {
    if ($cond) { Write-Host "  PASS: $msg" -ForegroundColor Green }
    else { Write-Host "  FAIL: $msg" -ForegroundColor Red; $script:failed++ }
}

Write-Host "== GET /health"
$h = Invoke-RestMethod "$Base/health"
Assert ($h.status -eq "ok") "status=ok"
Assert ($h.db -eq "ok") "db=ok"

Write-Host "== POST /api/decisions (эталон 2 000 000 / 36 мес / 20%)"
$body = @{
    inputs = @{
        loan_amount        = 2000000
        loan_term_months   = 36
        interest_rate_pct  = 20
        revenue            = 800000
        expenses           = 600000
        reserve            = 500000
        revenue_growth_pct = 30
        expense_growth_pct = 15
        purpose            = "Закупка оборудования"
    }
} | ConvertTo-Json
$bytes = [System.Text.Encoding]::UTF8.GetBytes($body)
$d = Invoke-RestMethod "$Base/api/decisions" -Method Post `
    -ContentType "application/json; charset=utf-8" `
    -Headers @{ "X-Max-User-Id" = "777" } -Body $bytes

Assert ($d.credit.monthly_payment -eq 74327) "платёж 74327 (got $($d.credit.monthly_payment))"
Assert ($d.credit.total_paid -eq 2675772) "выплаты 2675772 (got $($d.credit.total_paid))"
Assert ($d.scenarios.Count -eq 3) "3 сценария"
Assert ($d.scenarios[0].key -eq "expected" -and $d.scenarios[0].cash_flow -eq 275673) "🟢 cf=275673 (got $($d.scenarios[0].cash_flow))"
Assert ($d.scenarios[1].key -eq "moderate" -and $d.scenarios[1].cash_flow -eq 171673) "🟡 cf=171673 (got $($d.scenarios[1].cash_flow))"
Assert ($d.scenarios[2].key -eq "negative" -and $d.scenarios[2].cash_flow -eq -36327) "🔴 cf=-36327 (got $($d.scenarios[2].cash_flow))"
Assert ($d.scenarios[2].reserve_months -eq 13) "резерв на 13 мес (got $($d.scenarios[2].reserve_months))"
Assert ($d.checklist.Count -eq 5) "чек-лист 5 пунктов"
Assert ($d.scenarios[0].consequences.Count -ge 1) "последствия заполнены"

Write-Host "== GET /api/decisions/{id}"
$g = Invoke-RestMethod "$Base/api/decisions/$($d.id)"
Assert ($g.id -eq $d.id) "id совпадает"
Assert ($g.credit.monthly_payment -eq 74327) "платёж из БД 74327"
Assert ($g.scenarios.Count -eq 3) "3 сценария из БД"
Assert ($g.checklist.Count -eq 5) "чек-лист восстановлен"

Write-Host "== GET /api/decisions (история по пользователю 777)"
$l = Invoke-RestMethod "$Base/api/decisions" -Headers @{ "X-Max-User-Id" = "777" }
Assert ($l.items.Count -ge 1) "история не пуста"
Assert ($l.items[0].cash_flows.negative -eq -36327) "cash_flows в списке"

Write-Host "== POST c ошибкой валидации"
try {
    $bad = @{ inputs = @{ loan_amount = -5; loan_term_months = 0; revenue = 0 } } | ConvertTo-Json
    Invoke-RestMethod "$Base/api/decisions" -Method Post -ContentType "application/json" `
        -Body ([System.Text.Encoding]::UTF8.GetBytes($bad)) | Out-Null
    Assert $false "ожидался 400"
} catch {
    $code = $_.Exception.Response.StatusCode.value__
    Assert ($code -eq 400) "400 на невалидных данных (got $code)"
}

Write-Host "== GET несуществующий id → 404"
try {
    Invoke-RestMethod "$Base/api/decisions/00000000-0000-0000-0000-000000000000" | Out-Null
    Assert $false "ожидался 404"
} catch {
    $code = $_.Exception.Response.StatusCode.value__
    Assert ($code -eq 404) "404 (got $code)"
}

function Assert-Status($action, $wantCode, $msg) {
    try { & $action | Out-Null; $script:result = 200 } catch { $script:result = $_.Exception.Response.StatusCode.value__ }
    Assert ($script:result -eq $wantCode) "$msg (got $($script:result))"
}

Write-Host "== Границы запросов"
Assert-Status { Invoke-RestMethod "$Base/api/decisions?limit=0" } 400 "limit=0 → 400"
Assert-Status { Invoke-RestMethod "$Base/api/decisions?limit=abc" } 400 "limit=abc → 400"
Assert-Status { Invoke-RestMethod "$Base/api/decisions/not-a-uuid" } 400 "битый id → 400"
Assert-Status {
    Invoke-RestMethod "$Base/api/decisions" -Method Post -ContentType "application/json" `
        -Headers @{ "X-Max-User-Id" = "abc" } -Body ([System.Text.Encoding]::UTF8.GetBytes($body))
} 400 "битый X-Max-User-Id → 400"
Assert-Status {
    $unknown = '{ "inputs": { "loan_amount": 1, "interest_rate": 20 } }'
    Invoke-RestMethod "$Base/api/decisions" -Method Post -ContentType "application/json" `
        -Body ([System.Text.Encoding]::UTF8.GetBytes($unknown))
} 400 "неизвестное поле → 400"
Assert-Status {
    $huge = '{ "inputs": { "purpose": "' + ('a' * 70000) + '" } }'
    Invoke-RestMethod "$Base/api/decisions" -Method Post -ContentType "application/json" `
        -Body ([System.Text.Encoding]::UTF8.GetBytes($huge))
} 413 "тело > 64 КиБ → 413"

Write-Host "== DELETE /api/decisions/{id}"
$toDelete = Invoke-RestMethod "$Base/api/decisions" -Method Post `
    -ContentType "application/json; charset=utf-8" `
    -Headers @{ "X-Max-User-Id" = "777" } -Body $bytes
try {
    Invoke-RestMethod "$Base/api/decisions/$($toDelete.id)" -Method Delete `
        -Headers @{ "X-Max-User-Id" = "777" } | Out-Null
    $del1 = 204
} catch { $del1 = $_.Exception.Response.StatusCode.value__ }
Assert ($del1 -eq 204) "первое удаление → 204 (got $del1)"
try {
    Invoke-RestMethod "$Base/api/decisions/$($toDelete.id)" -Method Delete `
        -Headers @{ "X-Max-User-Id" = "777" } | Out-Null
    $del2 = 200
} catch { $del2 = $_.Exception.Response.StatusCode.value__ }
Assert ($del2 -eq 404) "повторное удаление → 404 (got $del2)"
Assert-Status { Invoke-RestMethod "$Base/api/decisions/$($toDelete.id)" } 404 "после удаления GET → 404"

Write-Host ""
if ($failed -eq 0) { Write-Host "SMOKE OK" -ForegroundColor Green; exit 0 }
else { Write-Host "SMOKE FAILED: $failed" -ForegroundColor Red; exit 1 }

