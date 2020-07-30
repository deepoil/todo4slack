package todo

import (
	"fmt"
	"github.com/slack-go/slack"
	"strconv"
	"todo4slack/util"
)

func UpdateDeadLine(action *slack.BlockAction) slack.MsgOption {
	// read todoList xml
	orgTodoList := util.ReadS3()
	var taskName string
	for i, task := range orgTodoList.Tasks {
		if strconv.Itoa(task.Id) == action.ActionID {
			orgTodoList.Tasks[i].DeadLine = action.SelectedDate
			taskName = string(task.Name)
		}
	}

	// update todoList xml
	util.UploadS3(orgTodoList)

	// return message to slack
	text := slack.NewTextBlockObject(
		slack.MarkdownType,
		fmt.Sprintf("updated task\n*%s* : *%s*", taskName, action.SelectedDate),
		false,
		false)
	textSection := slack.NewSectionBlock(text, nil, nil)

	return slack.MsgOptionBlocks(textSection)
}
