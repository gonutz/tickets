set GOARCH=386

go build -ldflags="-s -w" -o "Create new Ticket.exe" new_ticket.go
if errorlevel 1 pause

go build -ldflags="-s -w" -o "View Tickets.exe" view_tickets.go
if errorlevel 1 pause

echo git log --diff-filter=D --summary ./*.txt ^& pause>"Show Closed Tickets.bat"
