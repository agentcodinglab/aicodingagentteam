@echo off
setlocal enabledelayedexpansion
for /f "delims=" %%a in ('more') do (
  set "line=%%a"
  echo !line! | findstr /C:"initialize" >nul && echo {"jsonrpc":"2.0","id":1,"result":{"protocolVersion":"2025-03-26"}}
  echo !line! | findstr /C:"session/new" >nul && echo {"jsonrpc":"2.0","id":2,"result":{"sessionId":"s-stub"}}
  echo !line! | findstr /C:"session/prompt" >nul && (echo {"jsonrpc":"2.0","method":"notifications/session/update","params":{"sessionId":"s-stub","update":{"type":"agent_message_chunk","content":"hello from acp stub"}}} & echo {"jsonrpc":"2.0","id":3,"result":{"stopReason":"end_turn"}})
)
exit /b 0
