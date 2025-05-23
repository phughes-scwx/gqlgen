package chat

import (
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/99designs/gqlgen/client"
	"github.com/99designs/gqlgen/graphql/handler"
)

func TestChatSubscriptions(t *testing.T) {
	srv := handler.New(NewExecutableSchema(New()))
	srv.AddTransport(transport.Websocket{
		KeepAlivePingInterval: time.Second,
	})
	srv.AddTransport(transport.POST{})
	c := client.New(srv)

	const batchSize = 128
	var wg sync.WaitGroup
	for i := 0; i < batchSize*8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			sub := c.Websocket(fmt.Sprintf(
				`subscription @user(username:"vektah") { messageAdded(roomName:"#gophers%d") { text createdBy } }`,
				i,
			))
			defer sub.Close()

			var msg struct {
				resp struct {
					MessageAdded struct {
						Text      string
						CreatedBy string
					}
				}
				err error
			}

			msg.err = sub.Next(&msg.resp)
			if !assert.NoError(t, msg.err, "sub.Next") {
				return
			}
			assert.Equal(t, "You've joined the room", msg.resp.MessageAdded.Text)
			assert.Equal(t, "system", msg.resp.MessageAdded.CreatedBy)

			go func() {
				var resp any
				err := c.Post(fmt.Sprintf(`mutation {
					a:post(text:"Hello!", roomName:"#gophers%d", username:"vektah") { id }
					b:post(text:"Hello Vektah!", roomName:"#gophers%d", username:"andrey") { id }
					c:post(text:"Whats up?", roomName:"#gophers%d", username:"vektah") { id }
				}`, i, i, i), &resp)
				assert.NoError(t, err)
			}()

			msg.err = sub.Next(&msg.resp)
			if !assert.NoError(t, msg.err, "sub.Next") {
				return
			}
			assert.Equal(t, "Hello!", msg.resp.MessageAdded.Text)
			assert.Equal(t, "vektah", msg.resp.MessageAdded.CreatedBy)

			msg.err = sub.Next(&msg.resp)
			if !assert.NoError(t, msg.err, "sub.Next") {
				return
			}
			assert.Equal(t, "Whats up?", msg.resp.MessageAdded.Text)
			assert.Equal(t, "vektah", msg.resp.MessageAdded.CreatedBy)
		}(i)
		// wait for goroutines to finish every N tests to not starve on CPU
		if (i+1)%batchSize == 0 {
			wg.Wait()
		}
	}
	wg.Wait()

	// 1 for the main thread, 1 for the testing package and remainder is reserved for the HTTP server threads
	// TODO: use something like runtime.Stack to filter out HTTP server threads,
	// TODO: which is required for proper concurrency and leaks testing
	require.Less(t, runtime.NumGoroutine(), 1+1+batchSize*2, "goroutine leak")
}
