package todo

import (
	"github.com/slack-go/slack"
	"todo4slack/util"
)

func Done(id int) slack.MsgOption {
	// read todoList xml
	orgTodoList := util.ReadS3()
	var isExistTask = false
	var doneTaskName string

	for i, task := range orgTodoList.Tasks {
		if id == task.Id {
			doneTaskName = string(task.Name)
			orgTodoList.Tasks = append(orgTodoList.Tasks[:i], orgTodoList.Tasks[i+1:]...)
			isExistTask = true
			break
		}
	}

	var returnMessage string
	if isExistTask {
		// output xml-file
		util.UploadS3(orgTodoList)
		returnMessage = "Completed that *" + doneTaskName + "*"
	} else {
		returnMessage = "*That task has already been completed*\n" +
			"Get the latest task with this command [/td_show [options]] and check"
	}

	text := slack.NewTextBlockObject(slack.MarkdownType, returnMessage, false, false)
	textSection := slack.NewSectionBlock(text, nil, nil)
	return slack.MsgOptionBlocks(textSection)
}
