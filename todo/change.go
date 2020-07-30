package todo

import (
	"github.com/slack-go/slack"
	"strconv"
	"todo4slack/util"
)

func Change(id int) slack.MsgOption {
	// read todoList xml
	orgTodoList := util.ReadS3()
	taskFunc := func() *util.Task {
		for _, task := range orgTodoList.Tasks {
			if id == task.Id {
				return &util.Task{
					Id:       task.Id,
					Name:     task.Name,
					DeadLine: task.DeadLine,
				}
			}
		}
		return nil
	}
	var task = taskFunc()
	if task != nil {
		datePicker := &slack.DatePickerBlockElement{
			Type:        slack.METDatepicker,
			ActionID:    strconv.Itoa(task.Id),
			Placeholder: nil,
			InitialDate: task.DeadLine,
			Confirm:     nil,
		}
		// create message
		pickerText := slack.NewTextBlockObject(slack.PlainTextType, "change a deadline for this task", false, false)
		// set datePicker-section
		pickerSection := slack.NewSectionBlock(pickerText, nil, slack.NewAccessory(datePicker))
		// set "do not set" button
		denyButtonText := slack.NewTextBlockObject(slack.PlainTextType, "do not set", false, false)
		denyButton := slack.NewButtonBlockElement("", "deny", denyButtonText)
		denyButton.WithStyle(slack.StyleDanger)
		actionSection := slack.NewActionBlock(util.DoNotSetAction, denyButton)

		return slack.MsgOptionBlocks(pickerSection, actionSection)
	} else {
		returnMessage := "*That task has already been completed*\n" +
			"Get the latest task with this command [show <option>] and check"
		text := slack.NewTextBlockObject(slack.MarkdownType, returnMessage, false, false)
		textSection := slack.NewSectionBlock(text, nil, nil)
		return slack.MsgOptionBlocks(textSection)
	}

}
