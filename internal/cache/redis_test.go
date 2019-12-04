package cache

import (
	"encoding/json"
	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"kumparan/internal/config"
	"kumparan/internal/contract"
	"os"
	"testing"
	"time"
)

type RedisSuite struct {
	suite.Suite
	Host     string
	Password string
	DB       int
	Client   *redis.Client
}

func (r *RedisSuite) SetupSuite() {
	cfg := config.GetConfig()
	r.Client = redis.NewClient(&redis.Options{
		Addr:     r.Host,
		Password: cfg.Redis.Password,
		DB:       r.DB,
	})
}

func (r *RedisSuite) TearDownSuite() {
	_ = r.Client.Close()
}

type redisHandlerSuite struct {
	RedisSuite
}

func TestRedisSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skip test for redis repository")
	}
	redisHostTest := os.Getenv("REDIS_TEST_URL")
	if redisHostTest == "" {
		redisHostTest = "localhost:6379"
	}
	redisHandlerSuiteTest := &redisHandlerSuite{
		RedisSuite{
			Host: redisHostTest,
		},
	}
	suite.Run(t, redisHandlerSuiteTest)
}

func getItemByKey(client *redis.Client, key string) ([]byte, error) {
	return client.Get(key).Bytes()
}
func seedItem(client *redis.Client, key string, value interface{}) error {
	jybt, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return client.Set(key, jybt, time.Second*30).Err()
}
func (r *redisHandlerSuite) TestSet() {
	testkeynews := "news"
	repo := NewHandler(r.Client)

	str := "2014-11-12T11:45:26.371Z"
	t, _ := time.Parse(time.RFC3339, str)
	news := contract.NewsData{
		ID:      1,
		Author:  "bambang",
		Body:    "news",
		Created: t,
	}
	newsbyte, _ := json.Marshal(news)

	err := repo.Set(testkeynews, newsbyte, 1)
	require.NoError(r.T(), err)

	jbyt, err := getItemByKey(r.Client, testkeynews)
	require.NoError(r.T(), err)
	require.NotNil(r.T(), jbyt)
	var insertedData contract.NewsData
	err = json.Unmarshal(jbyt, &insertedData)
	require.NoError(r.T(), err)
	assert.Equal(r.T(), news.ID, insertedData.ID)
	assert.Equal(r.T(), news.Author, insertedData.Author)
	assert.Equal(r.T(), news.Body, insertedData.Body)
	assert.Equal(r.T(), news.Created, insertedData.Created)
}

func (r *redisHandlerSuite) TestGet() {
	testKeyName := "news"
	bola := contract.NewsData{Author: "Agus", Body: "anoa kidal", ID: 2}
	err := seedItem(r.Client, testKeyName, bola)
	require.NoError(r.T(), err)

	repo := NewHandler(r.Client)
	jbyt, err := repo.Get(testKeyName)
	require.NoError(r.T(), err)
	var res contract.NewsData
	err = json.Unmarshal(jbyt, &res)
	require.NoError(r.T(), err)

	assert.Equal(r.T(), bola.ID, res.ID)
	assert.Equal(r.T(), bola.Author, res.Author)
	assert.Equal(r.T(), bola.Body, res.Body)
}
