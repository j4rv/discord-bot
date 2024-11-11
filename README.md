# J4RV's Discord bot

[Add it to your server!](https://discord.com/oauth2/authorize?client_id=901475699699875880)

## How to run:

 - Needs Golang 1.20+
 - Windows: 
   - Needs GCC present in your path (For example https://jmeubank.github.io/tdm-gcc/)
   - Needs CGO_ENABLED=1 (`go env -w CGO_ENABLED=1`)

```
go run ./cmd/jarvbot -token **** -adminID ****
```

Example for ARM:
```
env CC=arm-linux-gnueabihf-gcc CXX=arm-linux-gnueabihf-g++ CGO_ENABLED=1 GOOS=linux GOARCH=arm GOARM=7 go run ./bin/ -token **** -adminID ****
```

## Available commands

https://github.com/j4rv/discord-bot/wiki/Help#available-commands