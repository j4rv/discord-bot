package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

const triviaBatchSize = "50"
const triviaTimeToWait = 15 * time.Second
const maxWinnersOnMessageLimitReached = 25
const triviaRateLimitCooldown = 30 * time.Second

var triviaRateLimitedUntil time.Time

// Distinct fruits for colour-blind friendlyness
var triviaAnswerEmojis = []string{
	"🍎", "🥥", "🍇", "🍉",
}

var triviaWinnerMessages = []string{
	"Congratulations to the winners!: %s",
	"Correct answers: %s",
	"5Heads detected: %s",
	"I did NOT expect these winners: %s",
	"The winners are: %s",
}

var triviaLoserMessages = []string{
	"The dummies that answered wrong: %s",
	"Wrong answers belong to: %s",
	"Better luck next time :3c (%s)",
	"Confidently incorrect: %s",
	"'At least you tried' reward goes to!: %s",
}

var triviaMultipleAnswerMessages = []string{
	"The extra dummies that chose more than one response: %s",
	"These cheaters couldn't pick just one: %s",
	"Indecisive dummies: %s",
	"You think you are real smart, huh?: %s",
	"PICK, JUST, ONE, PLEASE: %s",
}

type triviaCategory struct {
	ID   int
	Name string
}

var triviaCategories = map[string]triviaCategory{
	"general": {
		ID:   9,
		Name: "General Knowledge",
	},
	"books": {
		ID:   10,
		Name: "Entertainment: Books",
	},
	"film": {
		ID:   11,
		Name: "Entertainment: Film",
	},
	"music": {
		ID:   12,
		Name: "Entertainment: Music",
	},
	"musicals": {
		ID:   13,
		Name: "Entertainment: Musicals & Theatres",
	},
	"tv": {
		ID:   14,
		Name: "Entertainment: Television",
	},
	"videogames": {
		ID:   15,
		Name: "Entertainment: Video Games",
	},
	"boardgames": {
		ID:   16,
		Name: "Entertainment: Board Games",
	},
	"science": {
		ID:   17,
		Name: "Science & Nature",
	},
	"computers": {
		ID:   18,
		Name: "Science: Computers",
	},
	"math": {
		ID:   19,
		Name: "Science: Mathematics",
	},
	"mythology": {
		ID:   20,
		Name: "Mythology",
	},
	"sports": {
		ID:   21,
		Name: "Sports",
	},
	"geography": {
		ID:   22,
		Name: "Geography",
	},
	"history": {
		ID:   23,
		Name: "History",
	},
	"politics": {
		ID:   24,
		Name: "Politics",
	},
	"art": {
		ID:   25,
		Name: "Art",
	},
	"celebrities": {
		ID:   26,
		Name: "Celebrities",
	},
	"animals": {
		ID:   27,
		Name: "Animals",
	},
	"vehicles": {
		ID:   28,
		Name: "Vehicles",
	},
	"comics": {
		ID:   29,
		Name: "Entertainment: Comics",
	},
	"gadgets": {
		ID:   30,
		Name: "Science: Gadgets",
	},
	"anime": {
		ID:   31,
		Name: "Entertainment: Japanese Anime & Manga",
	},
	"cartoons": {
		ID:   32,
		Name: "Entertainment: Cartoon & Animations",
	},
}

func init() {
	commands["!trivia"] = notSpammable(answerTrivia)
	commands["!triviacredits"] = simpleTextResponse("Open Trivia DB: https://opentdb.com/")
}

type triviaQueryInput struct {
	Category   string `short:"c" long:"category" default:"" description:"Trivia category, or empty for any"`
	Difficulty string `short:"d" long:"difficulty" default:"" description:"Question difficulty: easy, medium, hard, or empty for any"`
	Type       string `short:"t" long:"type" default:"" description:"Question type: multiple, boolean, or empty for any"`
}

type validatedTriviaInput struct {
	Category   string
	CategoryID int
	Difficulty string
	Type       string
}

type openTDBResponse struct {
	ResponseCode int               `json:"response_code"`
	Results      []openTDBQuestion `json:"results"`
}

type openTDBQuestion struct {
	Category         string   `json:"category"`
	Type             string   `json:"type"`
	Difficulty       string   `json:"difficulty"`
	Question         string   `json:"question"`
	CorrectAnswer    string   `json:"correct_answer"`
	IncorrectAnswers []string `json:"incorrect_answers"`
}

func parseAndValidateTriviaInput(mc *discordgo.MessageCreate) (*validatedTriviaInput, string) {
	var input triviaQueryInput

	if err := parseCommandArgs(&input, mc.Content); err != nil {
		return nil, err.Error()
	}

	category := strings.ToLower(strings.TrimSpace(input.Category))
	categoryName, categoryID := "", 0
	if category != "" {
		triviaCategory, ok := triviaCategories[category]
		if !ok {
			categories := make([]string, 0, len(triviaCategories))
			for name := range triviaCategories {
				categories = append(categories, name)
			}
			sort.Strings(categories)
			return nil, fmt.Sprintf("Unknown trivia category: %s\nValid categories: %s", input.Category, strings.Join(categories, ", "))
		}
		categoryName = triviaCategory.Name
		categoryID = triviaCategory.ID
	}

	difficulty := strings.ToLower(strings.TrimSpace(input.Difficulty))
	switch difficulty {
	case "", "easy", "medium", "hard":
	default:
		return nil, "Difficulty must be easy, medium, hard, or empty."
	}

	questionType := strings.ToLower(strings.TrimSpace(input.Type))
	switch questionType {
	case "", "multiple", "boolean":
	default:
		return nil, "Type must be multiple, boolean, or empty."
	}

	return &validatedTriviaInput{
		Category:   categoryName,
		CategoryID: categoryID,
		Difficulty: difficulty,
		Type:       questionType,
	}, ""
}

func answerTrivia(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	validInput, errMsg := parseAndValidateTriviaInput(mc)
	if errMsg != "" {
		ds.ChannelMessageSend(mc.ChannelID, errMsg)
		return false
	}

	question, err := getTriviaQuestion(ctx, *validInput)
	if err != nil {
		adminNotifyIfErr("answerTrivia", err, ds)
		ds.ChannelMessageSend(mc.ChannelID, "Could not get a trivia question.")
		return false
	}
	if question == nil {
		ds.ChannelMessageSend(mc.ChannelID, "No trivia questions available.")
		return false
	}

	answers := []string{question.CorrectAnswer, question.IncorrectAnswer1}
	if question.IncorrectAnswer2 != nil {
		answers = append(answers, *question.IncorrectAnswer2)
	}
	if question.IncorrectAnswer3 != nil {
		answers = append(answers, *question.IncorrectAnswer3)
	}

	correctAnswer := 0
	if question.Type == "boolean" {
		if answers[0] == "False" {
			answers[0], answers[1] = answers[1], answers[0]
			correctAnswer = 1
		}
	} else {
		rand.Shuffle(len(answers), func(i, j int) {
			answers[i], answers[j] = answers[j], answers[i]
			if i == correctAnswer {
				correctAnswer = j
			} else if j == correctAnswer {
				correctAnswer = i
			}
		})
	}

	var answersFormatted []string
	for i, answer := range answers {
		answersFormatted = append(answersFormatted, fmt.Sprintf("%s %s", triviaAnswerEmojis[i], answer))
	}

	message := fmt.Sprintf("## %s\n```%s```", question.Question, strings.Join(answersFormatted, "\n"))
	sentMessage, err := ds.ChannelMessageSend(mc.ChannelID, message)
	if err != nil {
		serverNotifyIfErr("Trivia message could not be sent", err, mc.GuildID, ds)
		return false
	}

	err = triviaDS.incrementTriviaQuestionTimesUsed(question.ID)
	if err != nil {
		adminNotifyIfErr("incrementTriviaQuestionTimesUsed", err, ds)
	}

	for i := range answers {
		if err := ds.MessageReactionAdd(mc.ChannelID, sentMessage.ID, triviaAnswerEmojis[i]); err != nil {
			adminNotifyIfErr("answerTrivia reaction", err, ds)
		}
	}

	go waitForQuestionAnswersAndRespond(ds, mc.ChannelID, sentMessage.ID, correctAnswer)
	return true
}

func waitForQuestionAnswersAndRespond(ds *discordgo.Session, channelID string, messageID string, correctAnswer int) {
	time.Sleep(triviaTimeToWait)
	result := checkQuestionAnswers(ds, channelID, messageID, correctAnswer)

	message, err := ds.ChannelMessage(channelID, messageID)
	if err != nil {
		adminNotifyIfErr("waitForQuestionAnswersAndRespond", err, ds)
		return
	}

	lines := []string{message.Content, "**Time's up!!**", fmt.Sprintf("The correct answer was: %s", triviaAnswerEmojis[correctAnswer])}

	if len(result.CorrectUsers) > 0 {
		lines = append(lines, fmt.Sprintf(triviaWinnerMessages[rand.Intn(len(triviaWinnerMessages))], userListToMentions(result.CorrectUsers)))
	}
	if len(result.IncorrectUsers) > 0 {
		lines = append(lines, fmt.Sprintf(triviaLoserMessages[rand.Intn(len(triviaLoserMessages))], userListToMentions(result.IncorrectUsers)))
	}
	if len(result.MultipleAnswerUsers) > 0 {
		lines = append(lines, fmt.Sprintf(triviaMultipleAnswerMessages[rand.Intn(len(triviaMultipleAnswerMessages))], userListToMentions(result.MultipleAnswerUsers)))
	}

	content := strings.Join(lines, "\n")

	// Check discord's message length before editing
	if len(content) > discordMaxMessageLength {
		winners := result.CorrectUsers
		if len(winners) > maxWinnersOnMessageLimitReached {
			winners = winners[:maxWinnersOnMessageLimitReached]
		}
		winnerMentions := userListToMentions(winners)
		if len(result.CorrectUsers) > maxWinnersOnMessageLimitReached {
			winnerMentions += "..."
		}
		content = strings.Join([]string{message.Content, "**Time's up!!**",
			fmt.Sprintf("The correct answer was: %s", triviaAnswerEmojis[correctAnswer]),
			fmt.Sprintf(triviaWinnerMessages[rand.Intn(len(triviaWinnerMessages))], winnerMentions)}, "\n")
		if len(content) > discordMaxMessageLength {
			content = string([]rune(content)[:discordMaxMessageLength])
		}
	}

	_, err = ds.ChannelMessageEdit(channelID, messageID, content)
	if err != nil {
		adminNotifyIfErr("waitForQuestionAnswersAndRespond", err, ds)
	}
}

type triviaAnswerResult struct {
	CorrectUsers        []*discordgo.User
	IncorrectUsers      []*discordgo.User
	MultipleAnswerUsers []*discordgo.User
}

func checkQuestionAnswers(ds *discordgo.Session, channelID string, messageID string, correctAnswer int) triviaAnswerResult {
	var result triviaAnswerResult

	usersByEmoji := make(map[string]map[string]*discordgo.User)

	for _, emoji := range triviaAnswerEmojis {
		users, err := ds.MessageReactions(channelID, messageID, emoji, 100, "", "")
		if err != nil {
			adminNotifyIfErr("checkQuestionAnswers", err, ds)
			continue
		}

		usersByEmoji[emoji] = make(map[string]*discordgo.User)

		for _, user := range users {
			// Ignore the bot's own reaction.
			if user.ID == ds.State.User.ID {
				continue
			}
			usersByEmoji[emoji][user.ID] = user
		}
	}

	// Build a map of users -> the trivia reactions they used.
	userReactions := make(map[string][]string)
	for _, emoji := range triviaAnswerEmojis {
		for userID := range usersByEmoji[emoji] {
			userReactions[userID] = append(
				userReactions[userID],
				emoji,
			)
		}
	}

	correctEmoji := triviaAnswerEmojis[correctAnswer]
	for userID, reactions := range userReactions {
		user := usersByEmoji[reactions[0]][userID]
		switch len(reactions) {
		case 1:
			if reactions[0] == correctEmoji {
				result.CorrectUsers = append(result.CorrectUsers, user)
			} else {
				result.IncorrectUsers = append(result.IncorrectUsers, user)
			}
		default:
			result.MultipleAnswerUsers = append(result.MultipleAnswerUsers, user)
		}
	}

	return result
}

func getTriviaQuestion(ctx context.Context, input validatedTriviaInput) (*TriviaQuestion, error) {
	question, err := triviaDS.getLeastUsedTriviaQuestion(input.Category, input.Type, input.Difficulty)
	if err != nil {
		return nil, err
	}

	// Unused cached question found
	if question != nil && question.TimesUsed == 0 {
		return question, nil
	}

	// All matching questions have been used, get new ones
	questions, err := fetchTriviaQuestions(ctx, input)
	if err != nil {
		return nil, err
	}

	for _, tq := range questions {
		triviaDS.addTriviaQuestion(tq)
	}

	return triviaDS.getLeastUsedTriviaQuestion(input.Category, input.Type, input.Difficulty)
}

func fetchTriviaQuestions(ctx context.Context, input validatedTriviaInput) ([]openTDBQuestion, error) {
	if time.Now().Before(triviaRateLimitedUntil) {
		return nil, fmt.Errorf("OpenTDB rate limit active, try again later")
	}

	params := url.Values{}
	params.Set("amount", triviaBatchSize)
	params.Set("encode", "url3986")

	if input.Category != "" {
		params.Set("category", strconv.Itoa(input.CategoryID))
	}
	if input.Difficulty != "" {
		params.Set("difficulty", input.Difficulty)
	}
	if input.Type != "" {
		params.Set("type", input.Type)
	}

	requestURL := "https://opentdb.com/api.php?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		triviaRateLimitedUntil = time.Now().Add(triviaRateLimitCooldown)
		return nil, fmt.Errorf("OpenTDB rate limit exceeded")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenTDB returned HTTP status %d", resp.StatusCode)
	}

	var result openTDBResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.ResponseCode != 0 {
		return nil, fmt.Errorf("OpenTDB returned response code %d", result.ResponseCode)
	}

	for i := range result.Results {
		if err := decodeTriviaQuestion(&result.Results[i]); err != nil {
			return nil, err
		}
	}

	return result.Results, nil
}

func decodeTriviaQuestion(question *openTDBQuestion) error {
	var err error
	question.Category, err = url.PathUnescape(question.Category)
	if err != nil {
		return err
	}
	question.Type, err = url.PathUnescape(question.Type)
	if err != nil {
		return err
	}
	question.Difficulty, err = url.PathUnescape(question.Difficulty)
	if err != nil {
		return err
	}
	question.Question, err = url.PathUnescape(question.Question)
	if err != nil {
		return err
	}
	question.CorrectAnswer, err = url.PathUnescape(question.CorrectAnswer)
	if err != nil {
		return err
	}
	for i, answer := range question.IncorrectAnswers {
		question.IncorrectAnswers[i], err = url.PathUnescape(answer)
		if err != nil {
			return err
		}
	}
	return nil
}
