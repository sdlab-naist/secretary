package app

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/chez-shanpu/secretary/constants"
	"github.com/chez-shanpu/secretary/pkg/slack"

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
	_ = viper.BindEnv("LAB_DB_NAME")
	_ = viper.BindEnv("LAB_SLACK_TOKEN")
	_ = viper.BindEnv("LAB_SLACK_CHANNEL")
	_ = viper.BindEnv("LAB_SLACK_COMING_CHANNEL")

	// required
	_ = runCmd.MarkFlagRequired("config-path")
}

func runServer(cmd *cobra.Command, args []string) {
	var err error

	// db
	dataSrcName := fmt.Sprintf("%s:%s@/%s", viper.Get("LAB_DB_USER"), viper.Get("LAB_DB_PASSWORD"), viper.Get("LAB_DB_NAME"))
	db, err = sqlx.Open("mysql", dataSrcName)
	if err != nil {
		log.Fatalf("[ERROR] sqlx.open: %s", err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatalf("[ERROR] db.Ping: %s", err)
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
			log.Printf("[ERROR] resterEvent: %s", err)
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
	slackUsername, err := convertSlackUsername(username)
	if err != nil {
		return err
	}
	if status == constants.LabEventCome {
		msgStr = fmt.Sprintf("おかえりなさいませ，@%s様！ \n今日も一日頑張りましょう！", slackUsername)
	} else {
		msgStr = fmt.Sprintf("お疲れ様でした，@%s様！ \n帰り道気をつけてくださいね！", slackUsername)
	}

	token := viper.GetString("LAB_SLACK_TOKEN")
	ch := viper.GetString("LAB_SLACK_CHANNEL")
	mi := slack.NewSlackMessageInfo(token, ch, msgStr)
	err = mi.PostMessage()
	if err != nil {
		return err
	}

	if status != constants.LabEventCome {
		return nil
	}
	msgStr = fmt.Sprintf("@%s様がいらっしゃいました．", slackUsername)
	ch = viper.GetString("LAB_SLACK_COMING_CHANNEL")
	mi = slack.NewSlackMessageInfo(token, ch, msgStr)
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

func convertSlackUsername(name string) (string, error) {
	configPath := viper.Get("secretary.lab.config")
	path, fileName := filepath.Split(configPath.(string))
	fileNameExt := filepath.Ext(fileName)
	fileName = fileName[0 : len(fileName)-len(fileNameExt)]

	viper.SetConfigName(fileName)
	viper.AddConfigPath(path)
	err := viper.ReadInConfig()
	if err != nil {
		return "", err
	}
	nameMap := viper.Get("name").(map[string]interface{})
	slackName, ok := nameMap[name]
	if ok {
		return slackName.(string), nil
	} else {
		return name, nil
	}

}
