package todo

import (
	"fmt"
	"github.com/slack-go/slack"
	"sort"
	"strconv"
	"strings"
	"time"
	"todo4slack/util"
)

type todoList util.TodoList

// markdown default set
var closingChar = "*"

func (t todoList) Len() int {
	return len(t.Tasks)
}

func (t todoList) Swap(i, j int) {
	t.Tasks[i], t.Tasks[j] = t.Tasks[j], t.Tasks[i]
}

func (t todoList) Less(i, j int) bool {
	return t.Tasks[i].DeadLine < t.Tasks[j].DeadLine
}

func Show(command, option string) slack.MsgOption {
	// read todoList xml
	orgTodoList := util.ReadS3()

	// deadLine sort asc
	sort.Sort(todoList{Tasks: orgTodoList.Tasks})
	var blocks = make([]slack.Block, 0, len(orgTodoList.Tasks))

	// delimiter
	divider := slack.NewDividerBlock()

	// get current time
	t := time.Now()

	switch option {
	case "all":
		// create header block
		blocks = append(blocks, generateHeaderTextSection("All task list"))

		var preDeadLine = "f"
		for _, task := range orgTodoList.Tasks {
			if preDeadLine != task.DeadLine {
				var deadLine string
				if task.DeadLine == "" {
					deadLine = "<no deadline set>"
				} else {
					deadLine = task.DeadLine
				}
				deadLineText := slack.NewTextBlockObject(slack.PlainTextType, deadLine, false, false)
				textSection := slack.NewSectionBlock(deadLineText, nil, nil)
				blocks = append(blocks, divider)
				blocks = append(blocks, textSection)
			}

			if task.DeadLine != "" && t.Format("2006-01-02") > task.DeadLine {
				closingChar = "```"
			} else {
				closingChar = "*"
			}
			taskText := slack.NewTextBlockObject(slack.MarkdownType, fmt.Sprintf("%s%s%s", closingChar, task.Name, closingChar), false, false)
			blocks = append(blocks, slack.NewSectionBlock(taskText, nil, slack.NewAccessory(generateButtonElement(task, command))))

			preDeadLine = task.DeadLine
		}

	case "today":
		blocks = append(blocks, generateHeaderTextSection("Today's task list"))

		for _, task := range orgTodoList.Tasks {
			if t.Format("2006-01-02") == task.DeadLine {
				blocks = append(blocks, divider)
				taskText := slack.NewTextBlockObject(slack.MarkdownType, fmt.Sprintf("%s%s%s", closingChar, task.Name, closingChar), false, false)
				blocks = append(blocks, slack.NewSectionBlock(taskText, nil, slack.NewAccessory(generateButtonElement(task, command))))
			}
		}

	case "1w", "2w", "3w":
		excludeW := strings.Replace(option, "w", "", 1)
		val, _ := strconv.Atoi(excludeW)
		blocks = append(blocks, generateHeaderTextSection(fmt.Sprintf("Within %s week task list", excludeW)))

		var preDeadLine = "f"
		for _, task := range orgTodoList.Tasks {
			if task.DeadLine != "" &&
				t.Format("2006-01-02") <= task.DeadLine &&
				task.DeadLine <= t.AddDate(0, 0, 7*val).Format("2006-01-02") {
				if preDeadLine != task.DeadLine {
					deadLineText := slack.NewTextBlockObject(slack.PlainTextType, task.DeadLine, false, false)
					textSection := slack.NewSectionBlock(deadLineText, nil, nil)
					blocks = append(blocks, divider)
					blocks = append(blocks, textSection)
				}
				taskText := slack.NewTextBlockObject(slack.MarkdownType, fmt.Sprintf("%s%s%s", closingChar, task.Name, closingChar), false, false)
				blocks = append(blocks, slack.NewSectionBlock(taskText, nil, slack.NewAccessory(generateButtonElement(task, command))))

				preDeadLine = task.DeadLine
			}
		}

	case "exp":
		blocks = append(blocks, generateHeaderTextSection("Expired task list"))

		var preDeadLine = "f"
		for _, task := range orgTodoList.Tasks {
			if task.DeadLine != "" && t.Format("2006-01-02") > task.DeadLine {
				if preDeadLine != task.DeadLine {
					deadLineText := slack.NewTextBlockObject(slack.PlainTextType, task.DeadLine, false, false)
					textSection := slack.NewSectionBlock(deadLineText, nil, nil)
					blocks = append(blocks, divider)
					blocks = append(blocks, textSection)
				}
				taskText := slack.NewTextBlockObject(slack.MarkdownType, fmt.Sprintf("%s%s%s", closingChar, task.Name, closingChar), false, false)
				blocks = append(blocks, slack.NewSectionBlock(taskText, nil, slack.NewAccessory(generateButtonElement(task, command))))

				preDeadLine = task.DeadLine
			}
		}
	}
	// no task
	if len(blocks) < 3 {
		noTaskText := slack.NewTextBlockObject(slack.PlainTextType, "No tasks", false, false)
		blocks = append(blocks, slack.NewSectionBlock(noTaskText, nil, nil))
	}

	return slack.MsgOptionBlocks(blocks...)
}

// generate command header text
func generateHeaderTextSection(headerText string) *slack.SectionBlock {
	text := slack.NewTextBlockObject(slack.MarkdownType, fmt.Sprintf("%s###%s###%s", closingChar, headerText, closingChar), false, false)
	return slack.NewSectionBlock(text, nil, nil)
}

// generate task done button
func generateButtonElement(task util.Task, command string) *slack.ButtonBlockElement {
	buttonName := "done"
	if command == "/td_chg" {
		buttonName = "change"
	}
	doneButtonText := slack.NewTextBlockObject(slack.PlainTextType, buttonName, false, false)
	doneButton := slack.NewButtonBlockElement(strconv.Itoa(task.Id), buttonName, doneButtonText)
	return doneButton.WithStyle(slack.StylePrimary)
}
