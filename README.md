
Заполнить buildVersion, buildDate, buildCommit можно так:
```
task build -- -ldflags "-X main.buildVersion=1.0 -X main.buildDate=2026-03-24 -X main.buildCommit=some" 
```
