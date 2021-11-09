# J4RV's Discord bot

[Add it to your server!](https://discord.com/api/oauth2/authorize?client_id=901475699699875880&permissions=412384290880&scope=bot)

## How to run:

Needs Golang 1.17+

```
go run ./bin/ -token **** -adminID ****
```

Example for ARM:
```
env CC=arm-linux-gnueabihf-gcc CXX=arm-linux-gnueabihf-g++ CGO_ENABLED=1 GOOS=linux GOARCH=arm GOARM=7 go run ./bin/ -token **** -adminID ****
```

## Available commands (v1.1)

### Public

- **!source**: Links to the bot's source code
- **!ayayaify [message]**: Ayayaifies your message
- **!remindme [99h 99m 99s] [message]**: Reminds you of the message after the specified time has passed
- **!genshinDailyCheckIn**: Will remind you to do the Genshin Daily Check-In
- **!genshinDailyCheckInStop**: The bot will stop reminding you to do the Genshin Daily Check-In
- **!parametricTransformer**: Will remind you to use the Parametric Transformer every 7 days. Use it again to reset the reminder
- **!parametricTransformerStop**: The bot will stop reminding you to use the Parametric Transformer
- **!randomAbyssLineup**: The bot will give you two random teams and some replacements. Have fun ¯\_(ツ)_/¯. Optional: Write 8+ character names separated by commas and the bot will only choose from those
- **!randomArtifact**: Generates a random Lv20 Genshin Impact artifact
- **!randomArtifactSet**: Generates five random Lv20 Genshin Impact artifacts
- **!randomDomainRun (set A) (set B)**: Generates two random Lv20 Genshin Impact artifacts from the input sets

### Admin only

- **!reboot**: Reboot the bot's system
- **!shutdown [99h 99m 99s]**: Shuts down the bot's system
- **!abortShutdown**: Aborts the bot's system shutdown
