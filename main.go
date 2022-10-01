package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type Card struct {
	Term       string `json:"term,omitempty"`
	Definition string `json:"definition,omitempty"`
	Mistakes   int    `json:"mistakes,omitempty"`
}

type Deck struct {
	Cards []Card `json:"cards"`
}

func New() *Deck {
	return &Deck{Cards: make([]Card, 0)}
}

func initLogs() []string {
	return make([]string, 0)
}

var deck *Deck
var logs []string
var importFileName *string
var exportFileName *string

func main() {
	deck = New()
	logs = initLogs()
	importFileName = new(string)
	exportFileName = new(string)

	if len(os.Args) == 2 || len(os.Args) == 3 {
		importFileName = flag.String("import_from", "", "IMPORT FILE NAME")
		exportFileName = flag.String("export_to", "", "EXPORT FILE NAME")

		flag.Parse()

		if *importFileName != "" {
			importAction(*importFileName)
		}
		if *exportFileName != "" {
			exportAction(*exportFileName)
		}
	}
	for input := ""; input != "exit"; {
		printToTerminal("Input the action (add, remove, import, export, ask, exit, log, hardest card, reset stats):")
		input = getInput()
		makeActionForInput(input)
	}
}

func getInput() string {
	var reader = bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	logs = append(logs, input)
	return input
}

func createCard(deck *Deck) *Card {
	printToTerminal("The card:")
	term := getInput()
	for checkIfCardExists(deck, term) {
		printMessage := fmt.Sprintf(`The card "%s" already exists. Try again:`, term)
		printToTerminal(printMessage)
		term = getInput()
	}
	printToTerminal("The definition of the card:")
	definition := getInput()
	for checkIfCardExists(deck, definition) {
		printMessage := fmt.Sprintf(`The definition "%s" already exists. Try again:`, definition)
		printToTerminal(printMessage)
		definition = getInput()
	}
	return &Card{Term: term, Definition: definition}
}

func checkIfCardExists(deck *Deck, criteria string) bool {
	for _, card := range deck.Cards {
		if card.Term == criteria || card.Definition == criteria {
			return true
		}
	}
	return false
}

func readDeck(fileName string) (Deck, error) {
	var newDeck Deck
	file, err := os.ReadFile(fileName)
	if err != nil {
		printToTerminal("File not found.")
		return Deck{}, err
	}
	err = json.Unmarshal(file, &newDeck)
	if err != nil {
		log.Fatalln(err)
	}
	return newDeck, nil
}

func readFromFile(fileName string) *os.File {
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatalln(err)
	}
	return file
}

func printToTerminal(s string) {
	logs = append(logs, s)
	fmt.Println(s)
}

func saveInFile(fileName string) {
	file := readFromFile(fileName)
	defer file.Close()

	var contents []byte
	var newDeck *Deck

	if info, _ := file.Stat(); info.Size() > 0 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(file)
		contents = buf.Bytes()
		json.Unmarshal(contents, &newDeck)
	}

	if newDeck != nil && len(newDeck.Cards) > 0 {
		deck.Cards = append(deck.Cards, newDeck.Cards...)
	}

	res, err := json.Marshal(deck)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Fprintln(file, string(res))
}

func removeCard(term string) {
	for i, card := range deck.Cards {
		if card.Term == term {
			deck.Cards = append(deck.Cards[:i], deck.Cards[i+1:]...)
		}
	}
	printToTerminal("The card has been removed.\n")
}

func removeAction() {
	printToTerminal("Which card?")
	term := getInput()
	if checkIfCardExists(deck, term) {
		removeCard(term)
	} else {
		printToTerminal(fmt.Sprintf(`Can't remove "%s": there is no such card.`, term))
	}
}

func addAction() {
	card := createCard(deck)
	deck.Cards = append(deck.Cards, *card)
	printMessage := fmt.Sprintf(`The pair ("%s":"%s") has been added.`, card.Term, card.Definition)
	printToTerminal(printMessage)
}

func importAction(fileName string) {
	newDeck, err := readDeck(fileName)
	if err != nil {
		return
	}
	printMessage := fmt.Sprintf("%d cards have been loaded.", len(newDeck.Cards))
	printToTerminal(printMessage)
	deck.Cards = append(deck.Cards, newDeck.Cards...)
}

func exportAction(fileName string) {
	saveInFile(fileName)
	printMessage := fmt.Sprintf("%d cards have been saved.", len(deck.Cards))
	printToTerminal(printMessage)
}

func getFileName() string {
	printToTerminal("File name:")
	fileName := getInput()
	return fileName
}

func askAction() {
	times := getTimesToAsk()
	if len(deck.Cards) == 0 {
		printToTerminal("Not enough cards...")
		return
	}
outer:
	for i := 0; i < times; i++ {
		var index int = i
		if i >= len(deck.Cards) {
			index = i % len(deck.Cards)
		}
		card := &deck.Cards[index]
		printMessage := fmt.Sprintf(`Print the definition of "%s":`, card.Term)
		printToTerminal(printMessage)
		answer := getInput()
		if answer == card.Definition {
			printToTerminal("Correct!")
			continue
		}
		for _, fCard := range deck.Cards {
			if fCard.Definition == answer {
				card.Mistakes++
				printMessage = fmt.Sprintf(`Wrong. The right answer is "%s", but your definition is correct for "%s".`, card.Definition, fCard.Term)
				printToTerminal(printMessage)
				continue outer
			}
		}
		card.Mistakes++
		printMessage = fmt.Sprintf(`Wrong. The right answer is "%s".`, card.Definition)
		printToTerminal(printMessage)
	}
}

func getTimesToAsk() int {
	printToTerminal("How many times to ask?")
	numberOfCards, _ := strconv.Atoi(getInput())
	return numberOfCards
}

func logAction() {
	fileName := getFileName()
	file := readFromFile(fileName)
	defer file.Close()
	l, _ := json.Marshal(logs)
	fmt.Fprintln(file, string(l))
	printToTerminal("The log has been saved.")
}

func hardestCardAction() {
	if len(deck.Cards) == 0 {
		printToTerminal("There are no cards with errors.")
		return
	}
	var highestMistakes int
	mistakes := make(map[string]Card, 0)
	for _, card := range deck.Cards {
		if card.Mistakes > highestMistakes {
			highestMistakes = card.Mistakes
		}
	}
	if highestMistakes == 0 {
		printToTerminal("There are no cards with errors.")
		return
	}
	for _, card := range deck.Cards {
		if card.Mistakes == highestMistakes {
			mistakes[card.Term] = card
		}
	}
	if len(mistakes) == 1 {
		keys := make([]string, 0, len(mistakes))
		for k := range mistakes {
			keys = append(keys, k)
		}
		key := keys[0]

		printMessage := fmt.Sprintf(`The hardest card is "%s". You have %d errors answering it.`, mistakes[key].Term, mistakes[key].Mistakes)
		printToTerminal(printMessage)
		return
	}
	var printMistakes string
	var mistakesCount int
	for _, mistake := range mistakes {
		printMistakes += fmt.Sprintf(`"%s", `, mistake.Term)
		mistakesCount = mistake.Mistakes
	}
	printMistakes = printMistakes[:len(printMistakes)-1]
	printMessage := fmt.Sprintf(`The hardest cards are %s. You have %d errors answering them.`, printMistakes, mistakesCount)
	printToTerminal(printMessage)
}

func resetAction() {
	for i := range deck.Cards {
		deck.Cards[i].Mistakes = 0
	}
	printToTerminal("Card statistics have been reset.")
}

func makeActionForInput(input string) {
	switch input {
	case "add":
		addAction()
	case "remove":
		removeAction()
	case "import":
		importAction(getFileName())
	case "export":
		exportAction(getFileName())
	case "ask":
		askAction()
	case "log":
		logAction()
	case "hardest card":
		hardestCardAction()
	case "reset stats":
		resetAction()
	case "exit":
		if *importFileName != "" {
			importAction(*importFileName)
		}
		if *exportFileName != "" {
			exportAction(*exportFileName)
		}
		printToTerminal("Bye bye!")
	default:
		fmt.Printf("%s not implemented", input)
	}
}
