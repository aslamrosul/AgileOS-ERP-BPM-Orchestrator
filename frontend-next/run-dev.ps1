# Development script untuk menjalankan Next.js frontend

Write-Host "🚀 Starting AgileOS Frontend (Development Mode)" -ForegroundColor Green
Write-Host "   URL: http://localhost:3000" -ForegroundColor Cyan
Write-Host "   Backend: $env:NEXT_PUBLIC_API_URL" -ForegroundColor Cyan
Write-Host ""
Write-Host "⚠️  Make sure backend is running on port 8080!" -ForegroundColor Yellow
Write-Host ""

npm run dev
