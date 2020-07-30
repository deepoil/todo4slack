package todo

import (
	"github.com/slack-go/slack"
	"regexp"
	"strconv"
	"time"
	"todo4slack/util"
)

func AddTask(options []string) slack.MsgOption {
	var todoMessageSlice = make([]byte, 0, 100)
	for _, val := range options {
		val = regexp.MustCompile("\r\n|\r|\n").ReplaceAllString(val, " ")
		todoMessageSlice = append(todoMessageSlice, val...)
		todoMessageSlice = append(todoMessageSlice, ' ')
	}

	// read todoList xml
	orgTodoList := util.ReadS3()

	// get latest id
	latestId := util.GetLatestId(orgTodoList)

	// create new task struct
	task := &util.Task{Id: latestId, Name: todoMessageSlice}

	// create new todoList struct
	tasks := append(orgTodoList.Tasks, *task)
	newTodoList := &util.TodoList{Tasks: tasks}

	// output xml-file
	util.UploadS3(newTodoList)

	// result message to slack
	returnMessage := "created task : *" + string(task.Name) + "*"
	text := slack.NewTextBlockObject(slack.MarkdownType, returnMessage, false, false)
	textSection := slack.NewSectionBlock(text, nil, nil)

	// return datePicker to slack
	t := time.Now()
	datePicker := &slack.DatePickerBlockElement{
		Type:        slack.METDatepicker,
		ActionID:    strconv.Itoa(latestId),
		Placeholder: nil,
		InitialDate: t.Format("2006-01-02"),
		Confirm:     nil,
	}
	// create message
	pickerText := slack.NewTextBlockObject(slack.PlainTextType, "set a deadline for this task", false, false)
	// set datePicker-section
	pickerSection := slack.NewSectionBlock(pickerText, nil, slack.NewAccessory(datePicker))

	// set "do not set" button
	denyButtonText := slack.NewTextBlockObject(slack.PlainTextType, "do not set", false, false)
	denyButton := slack.NewButtonBlockElement("", "deny", denyButtonText)
	denyButton.WithStyle(slack.StyleDanger)
	actionSection := slack.NewActionBlock(util.DoNotSetAction, denyButton)

	// delimiter
	divider := slack.NewDividerBlock()
	return slack.MsgOptionBlocks(textSection, divider, pickerSection, actionSection)
}
