package app

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"sync"
	"text/template"
	"time"

	"github.com/chez-shanpu/secretary/pkg/user"

	"github.com/chez-shanpu/secretary/constants"
	myslack "github.com/chez-shanpu/secretary/pkg/slack"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run secretary-lab server",
	Long:  "run secretary-lab server",
	Run:   runServer,
}

var db *sqlx.DB

type ReqStatus struct {
	Username string `json:"name" binding:"required"`
}

func init() {
	rootCmd.AddCommand(runCmd)

	// flags
	flags := runCmd.Flags()
	flags.StringP("config-path", "p", "", "path to yaml config for binding username and slack_username")

	// bind flags
	_ = viper.BindPFlag("secretary.lab.config", flags.Lookup("config-path"))

	// bind env vars
	_ = viper.BindEnv("LAB_DB_USER")
	_ = viper.BindEnv("LAB_DB_PASSWORD")
	_ = viper.BindEnv("LAB_DB_PROTOCOL")
	_ = viper.BindEnv("LAB_DB_ADDR")
	_ = viper.BindEnv("LAB_DB_NAME")
	_ = viper.BindEnv("LAB_SLACK_TOKEN")
	_ = viper.BindEnv("LAB_SLACK_COMING_CHANNEL")

	// required
	_ = runCmd.MarkFlagRequired("config-path")
}

func runServer(cmd *cobra.Command, args []string) {
	var err error

	// db
	dataSrcName := fmt.Sprintf("%s:%s@%s(%s)/%s",
		viper.Get("LAB_DB_USER"),
		viper.Get("LAB_DB_PASSWORD"),
		viper.Get("LAB_DB_PROTOCOL"),
		viper.Get("LAB_DB_ADDR"),
		viper.Get("LAB_DB_NAME"))
	db, err = sqlx.Open("mysql", dataSrcName)
	if err != nil {
		log.Fatalf("[ERROR] sqlx.open: %s", err)
	}
	for {
		err = db.Ping()
		if err != nil {
			log.Printf("[Info] db.Ping: %s", err)
			time.Sleep(5 * time.Second)
		} else {
			break
		}
	}

	defer db.Close()

	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.GET("/status", getStatus)
	r.POST("/event", postEvent)

	r.Run(":8080")
}

func getStatus(c *gin.Context) {
	username := c.Query("name")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "please input name parameter"})
		return
	}

	status, err := getCurrentStatus(username)
	if err != nil {
		log.Printf("[ERROR] getCurrentStatus db GET: %s", err)
		c.Status(http.StatusInternalServerError)
		return
	}

	var msg string
	if status == constants.LabEventCome {
		msg = fmt.Sprintf("%s is comming!", username)
	} else {
		msg = fmt.Sprintf("%s has left...", username)
	}
	c.JSON(http.StatusOK, gin.H{"message": msg})
}

func postEvent(c *gin.Context) {
	var req ReqStatus
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	status, err := getCurrentStatus(req.Username)
	if err != nil {
		log.Printf("[ERROR] getCurrentStatus db GET: %s", err)
		c.Status(http.StatusInternalServerError)
		return
	}

	var nowStatus string
	if status == constants.LabEventCome {
		nowStatus = constants.LabEventLeave
	} else {
		nowStatus = constants.LabEventCome
	}

	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func(u, s string) {
		err := regesterEvent(u, s)
		if err != nil {
			log.Printf("[ERROR] regesterEvent: %s", err)
		}
		wg.Done()
	}(req.Username, nowStatus)
	go func(u, s string) {
		err := sendMessage(u, s)
		if err != nil {
			log.Printf("[ERROR] sendMessage: %s", err)
		}
		wg.Done()
	}(req.Username, nowStatus)
	wg.Wait()

	c.Status(http.StatusOK)
}

func regesterEvent(username, status string) error {
	res, err := db.Exec("INSERT INTO lab_events (`username`, `event_type`) VALUES (?,?)", username, status)
	log.Printf("[INFO] regesterEvent query reseult: %s", res)
	return err
}

func sendMessage(username, status string) error {
	var msgStr string
	u, err := user.GetUser(viper.GetString("secretary.lab.config"), username)
	if err != nil {
		return err
	}

	var t *template.Template
	if status == constants.LabEventCome {
		tmpl := u.SecretaryComingMsg
		t, err = template.New("coming").Parse(tmpl)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		tmpl := u.SecretaryGoodbyeMsg
		t, err = template.New("goodbye").Parse(tmpl)
		if err != nil {
			log.Fatal(err)
		}
	}

	buf := new(bytes.Buffer)
	if err = t.Execute(buf, u); err != nil {
		log.Fatal(err)
	}
	msgStr = buf.String()

	token := viper.GetString("LAB_SLACK_TOKEN")

	var ch string
	var mi *myslack.MessageInfo
	if u.SlackChannel != "nil" {
		ch = u.SlackChannel
		mi = myslack.NewSlackMessageInfo(token, ch, u.SecretaryName, u.SecretaryIcon, msgStr)
		err = mi.PostMessage()
		if err != nil {
			return err
		}
	}
	if status != constants.LabEventCome {
		return nil
	}
	msgStr = fmt.Sprintf("<@%s>くんが来ました．", u.SlackId)
	ch = viper.GetString("LAB_SLACK_COMING_CHANNEL")
	mi = myslack.NewSlackMessageInfo(token, ch, u.SecretaryName, u.SecretaryIcon, msgStr)
	err = mi.PostMessage()
	return err
}

func getCurrentStatus(username string) (string, error) {
	var todayEventNum int

	err := db.Get(&todayEventNum, "SELECT count(*) FROM lab_events WHERE username=? AND created_at > CURRENT_DATE", username)
	if err != nil {
		return "", err
	}

	if todayEventNum%2 == 0 {
		return constants.LabEventLeave, nil
	} else {
		return constants.LabEventCome, nil
	}
}
