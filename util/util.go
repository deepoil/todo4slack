package util

import (
	"encoding/xml"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/slack-go/slack"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"todo4slack/config"
)

const (
	FileName            = "_TodoForSlack.xml"
	PickerVersionAction = "datepicker"
	DoNotSetAction      = "do-not-button"
)

var (
	UserId string
	TeamId string
)

type TodoList struct {
	Tasks []Task `xml:"Task"`
}

type Task struct {
	Id       int
	Name     []byte
	DeadLine string
}

func Usage() slack.MsgOption {
	description :=
		`usage: 
	*Add a task to the to do list*
		/td_add <task name>


	*Change task due date*
		/td_chg [options]

			[all]		Show all todo list
			[today]		Display todo list for today
			[1w]		Display todo list within 1 week
			[2w]		Display todo list within 2 week
			[3w]		Display todo list within 3 week
			[exp]		Show expired todo list


	*Display the list. You can also complete tasks from this command*
		/td_show [options]

			[all]		Show all todo list
			[today]		Display todo list for today
			[1w]		Display todo list within 1 week
			[2w]		Display todo list within 2 week
			[3w]		Display todo list within 3 week
			[exp]		Show expired todo list
`
	text := slack.NewTextBlockObject(slack.MarkdownType, description, false, false)
	textSection := slack.NewSectionBlock(text, nil, nil)
	return slack.MsgOptionBlocks(textSection)
}

func connectionAws() (*session.Session, error) {
	creds := credentials.NewStaticCredentials(config.Config.AwsAccessKey, config.Config.AwsSecretKey, "")
	sess, err := session.NewSession(&aws.Config{
		Credentials: creds,
		Region:      aws.String("ap-northeast-1"),
	})
	if err != nil {
		return nil, err
	}
	return sess, nil
}

func UploadS3(todoList *TodoList) {
	// write local todoFile
	todoFile, err := writeTodoFile(todoList)
	if err != nil {
		log.Fatal(err)
	}

	sess, err := connectionAws()
	if err != nil {
		log.Fatal(err)
	}

	file, err := os.Open(todoFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	objectKey := TeamId + "/" + UserId + FileName
	uploader := s3manager.NewUploader(sess)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(config.Config.BucketName),
		Key:    aws.String(objectKey),
		Body:   file,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func ReadS3() *TodoList {
	sess, err := connectionAws()
	if err != nil {
		log.Fatal(err)
	}
	var todoList TodoList
	objectKey := TeamId + "/" + UserId + FileName
	svc := s3.New(sess)
	obj, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(config.Config.BucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return &todoList
	}

	rc := obj.Body
	defer rc.Close()

	data, err := ioutil.ReadAll(rc)
	if err != nil {
		log.Fatal(err)
	}

	if err := xml.Unmarshal(data, &todoList); err != nil {
		log.Fatal("unmarshal error: ", err)
	}

	return &todoList
}

func GenerateTodoDirInfo() string {
	return filepath.Join(os.Getenv("HOME"), "todoSlack")
}

func writeTodoFile(todoList *TodoList) (string, error) {
	// create directory
	dir := GenerateTodoDirInfo()
	_ = os.Mkdir(dir, 0777)

	// open file
	file, err := os.OpenFile(filepath.Join(dir, TeamId+"_"+UserId+FileName), os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0777)
	if err != nil {
		log.Fatal("init file error: ", err)
	}
	defer file.Close()

	// convert struct to xml
	buf, _ := xml.MarshalIndent(todoList, "", " ")

	// output file
	_, err = fmt.Fprintln(file, string(buf))
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	return file.Name(), nil
}

func GetLatestId(t *TodoList) int {
	// order by Id desc
	sortedList := OrderByIdDesc(t)

	// generate new id
	var latestId int
	if len(sortedList.Tasks) == 0 {
		latestId = 1
	} else {
		latestId = sortedList.Tasks[0].Id + 1
	}
	return latestId
}

func OrderByIdDesc(orgTodoList *TodoList) TodoList {
	todoList := TodoList{
		Tasks: make([]Task, len(orgTodoList.Tasks)),
	}
	copy(todoList.Tasks, orgTodoList.Tasks)
	sort.Sort(todoList)
	return todoList
}

func (t TodoList) Len() int {
	return len(t.Tasks)
}

func (t TodoList) Swap(i, j int) {
	t.Tasks[i], t.Tasks[j] = t.Tasks[j], t.Tasks[i]
}

func (t TodoList) Less(i, j int) bool {
	return t.Tasks[i].Id > t.Tasks[j].Id
}
