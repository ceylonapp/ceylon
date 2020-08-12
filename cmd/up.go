package cmd

import (
	"ceylon/cli/mgt/virtualenv"
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/cobra"
	"log"
	"os"
)

var up = &cobra.Command{
	Use: "up",
	Run: func(cmd *cobra.Command, args []string) {
		var ctx = context.Background()

		rdb := redis.NewClient(&redis.Options{
			Addr: "192.168.8.100:6379",
			DB:   0, // use default DB
		})
		defer rdb.Close()

		pubsub := rdb.Subscribe(ctx, "ceylon_chess_player_sys_log")

		messages := make(chan *redis.Message)

		forceCreate, err := cmd.Flags().GetBool("forceCreate")
		if err != nil {
			panic(err)
		}
		deployManager := virtualenv.CreateInstance(context.Background())

		err = deployManager.Create(&virtualenv.CreateSettings{ForceCreate: forceCreate})
		err = deployManager.Prepare()
		err = os.Chdir(deployManager.ProjectPath)

		if err != nil {
			panic(err)
		}
		go func() {
			log.Print("Agents starting...")
			deployManager.Run()
		}()
		fmt.Println("All Agents are up and running")

		go func() {
			msg, err := pubsub.ReceiveMessage(ctx)
			if err != nil {
				panic(err)
			}
			//fmt.Println(msg.Channel, msg.Payload)
			messages <- msg
		}()

		for {
			msg := <-messages
			fmt.Println("Received message ", msg.Channel, msg.Pattern, msg.Payload)
		}
	},
}
