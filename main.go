package main

import (
	"bytes"
	"encoding/json"
	"github.com/slack-go/slack"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"todo4slack/config"
	"todo4slack/todo"
	"todo4slack/util"
)

// command event
func commandHandler(w http.ResponseWriter, r *http.Request) {
	api := slack.New(config.Config.ApiToken)

	commands, err := slack.SlashCommandParse(r)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	util.TeamId = commands.TeamID
	util.UserId = commands.UserID
	options := strings.Split(commands.Text, " ")
	var blocks slack.MsgOption
	switch commands.Command {
	case "/td_show", "/td_chg":
		if strings.TrimSpace(options[0]) == "" {
			text := slack.NewTextBlockObject(slack.MarkdownType, "Please enter the option", false, false)
			textSection := slack.NewSectionBlock(text, nil, nil)
			blocks = slack.MsgOptionBlocks(textSection)
		} else {
			blocks = todo.Show(commands.Command, options[len(options)-1])
		}

		if _, err := api.PostEphemeral(commands.ChannelID, commands.UserID, blocks); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

	case "/td_add":
		if strings.TrimSpace(options[0]) == "" {
			text := slack.NewTextBlockObject(slack.MarkdownType, "Please enter the task name", false, false)
			textSection := slack.NewSectionBlock(text, nil, nil)
			blocks = slack.MsgOptionBlocks(textSection)
		} else {
			blocks = todo.AddTask(options[:])
		}

		if _, err := api.PostEphemeral(commands.ChannelID, commands.UserID, blocks); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	case "/td_usage":
		blocks = util.Usage()
		if _, err := api.PostEphemeral(commands.ChannelID, commands.UserID, blocks); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

// slack api action
func actionHandler(w http.ResponseWriter, r *http.Request) {
	api := slack.New(config.Config.ApiToken)
	verifier, err := slack.NewSecretsVerifier(r.Header, config.Config.ApiSecret)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	bodyReader := io.TeeReader(r.Body, &verifier)
	body, err := ioutil.ReadAll(bodyReader)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := verifier.Ensure(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	var payload *slack.InteractionCallback
	if err := json.Unmarshal([]byte(r.FormValue("payload")), &payload); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	util.TeamId = payload.Team.ID
	util.UserId = payload.User.ID
	switch payload.Type {
	case slack.InteractionTypeBlockActions:
		if len(payload.ActionCallback.BlockActions) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		action := payload.ActionCallback.BlockActions[0]
		switch action.Type {
		// set deadline
		case util.PickerVersionAction:
			blocks := todo.UpdateDeadLine(action)
			fallbackText := slack.MsgOptionText("not found this task", false)
			replaceOriginal := slack.MsgOptionReplaceOriginal(payload.ResponseURL)
			if _, _, _, err := api.SendMessage("", replaceOriginal, fallbackText, blocks); err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

		case "button":
			// do not set deadline
			if action.Value == "deny" {
				replaceOriginal := slack.MsgOptionReplaceOriginal(payload.ResponseURL)
				text := slack.NewTextBlockObject(slack.MarkdownType, "*You can set deadline later!*", false, false)
				textSection := slack.NewSectionBlock(text, nil, nil)
				if _, _, _, err := api.SendMessage("", replaceOriginal, slack.MsgOptionBlocks(textSection)); err != nil {
					log.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				// task done
			} else if action.Value == "done" {
				responseOriginal := slack.MsgOptionResponseURL(payload.ResponseURL, "Post")
				actionId, _ := strconv.Atoi(action.ActionID)
				blocks := todo.Done(actionId)
				if _, _, _, err := api.SendMessage("", responseOriginal, blocks); err != nil {
					log.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				// task change
			} else if action.Value == "change" {
				responseOriginal := slack.MsgOptionResponseURL(payload.ResponseURL, "Post")
				actionId, _ := strconv.Atoi(action.ActionID)
				blocks := todo.Change(actionId)
				if _, _, _, err := api.SendMessage("", responseOriginal, blocks); err != nil {
					log.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}
		}
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/slack/command", commandHandler)
	http.HandleFunc("/slack/actions", actionHandler)
	log.Println("[INFO] Server listening")
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
