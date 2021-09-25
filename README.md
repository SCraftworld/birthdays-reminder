# birthdays-reminder
Minimalistic windows app which reminds about upcoming birthdays

## Why not calendar events?
1. Calendar requires additional (complex?) configuration in order to show age and to trigger at convenient time (which is a downside considered that this app designed for and used by elderly people);
2. There is a need to access all known birthday dates at once (e.g. group by location/scan over). Solution in the form of a single .txt file seems sufficient;
3. Not all dates are fully defined (e.g. exact day may be unknown). Such cases require special handling.

Aforementioned reasons interwoven with my personal desire to try golang has led to birth of this application.

## Build
go build -ldflags -H=windowsgui

## BD.txt example
```
01.01.2012 Name 1
01.01.2012 Name 2

//comment line
??.04.???? Name 3
```

## Installation
1. Put executable and BD.txt in same directory;
2. Run executable and allow elevated rights when asked (this is required once for autostart registration);
3. Done.