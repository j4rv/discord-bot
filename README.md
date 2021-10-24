# J4RV's Discord bot

## How to run:

Needs Golang 1.17+

```
go run ./bin/ -token **** -adminID ****
```

## Available commands (v1.1)

### Public

- **!genshinDailyCheckIn**: Will remind you to do the Genshin Daily Check-In
- **!genshinDailyCheckInStop**: The bot will stop reminding you to do the Genshin Daily Check-In
- **!parametricTransformer**: Will remind you to use the Parametric Transformer every 7 days
- **!parametricTransformerStop**: The bot will stop reminding you to use the Parametric Transformer
- **!ayayaify [message]**: Ayayaifies your message
- **!remindme [99h 99m 99s] [message]**: Reminds you of the message after the specified time has passed

### Admin only

- **!reboot**: Reboot the bot's system
- **!shutdown [99h 99m 99s]**: Shuts down the bot's system
- **!abortShutdown**: Aborts the bot's system shutdown
