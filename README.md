# Tickets for Windows

The way I handle tickets these days is by just adding a folder `tickets` to my root project folder and commit it to version control. The folder contains plain text files `1.txt`, `2.txt`, etc. one file per ticket.

Creating a new ticket means writing a new text file and committing it to version control. Change the text file and commit it to update that ticket. Delete the text file to mark the ticket as fixed, it will still be available in your repository history if you want to view it later.

Having tickets in your version control puts them right where you work, no need for an extra bug tracker, everybody working with the code has direct access to the tickets. Seeing what has been done by whom is as simple as using a git command to see who deleted which `.txt` files in the tickets folder. Other common tasks can be done in a similar way.

This repository contains Windows GUI apps, written in Go, to make common tasks easier when working in this fashion:

- `view_tickets.go` shows all tickets and allows easy search and ticket deletion
- `new_ticket.go` lets you create a ticket with a short title and detailed description

In order to build the tools you need to have [Go](https://golang.org/) installed.

Go into your project's root folder and run

    git clone https://github.com/gonutz/tickets & cd tickets & call build_ticket_apps.bat & for /F %f in ('git ls-tree -r master --name-only') do del "%f" & rmdir /S /Q .git & cd ..

from the command line. This will create a new folder `tickets` and build the GUI apps in it.