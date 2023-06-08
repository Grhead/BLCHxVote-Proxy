package main

import (
	"Vox2-Proxy/Transport"
	. "Vox2-Proxy/Transport/PBs"
	"context"
	"fmt"
	"github.com/gin-contrib/sessions"
	intersessions "github.com/gin-contrib/sessions/gorm"
	"github.com/gin-gonic/gin"
	googleProtobuf "github.com/golang/protobuf/ptypes/empty"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"net/http"
)

type Store interface {
	sessions.Store
}

var router *gin.Engine
var store Store
var grpcClient ContractClient

func main() {
	viper.SetConfigFile("./LowConf/config.env")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalln(err)
	}
	db, err := gorm.Open(sqlite.Open("Database/SessionsDb.db"), &gorm.Config{})
	if err != nil {
		log.Fatalln(err)
	}
	bindPort := viper.GetString("BIND_PORT")

	store = intersessions.NewStore(db, true, []byte("secret"))
	router = gin.Default()
	router.Use(sessions.Sessions("BaseSession", store))
	router.POST("/check", GinCheck)
	router.POST("/register", GinRegister)
	router.POST("/acceptLoad", GinAcceptLoadUser)
	router.POST("/acceptNew", GinAcceptNewUser)
	router.POST("/vote", GinVote)
	router.POST("/solo", GinSoloWinner)
	router.POST("/list", GinWinnersList)
	router.POST("/viewCandidates", GinViewCandidates)
	router.POST("/newCandidate", GinNewCandidates)
	router.POST("/createVoters", GinCallCreateVoters)
	router.POST("/getFull", GinGetFull)
	router.POST("/getPart", GinGetPartOfChain)
	router.POST("/size", GinGetChainSize)
	router.POST("/newChain", GinNewChain)
	conn, err := grpc.Dial(bindPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalln(err)
	}
	grpcClient = NewContractClient(conn)
	err = router.Run(":8199")
	if err != nil {
		log.Fatalln(err)
	}
}

func GinRegister(c *gin.Context) {
	var input *Transport.AuthStruct
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": err.Error()})
		return
	} else {
		session := sessions.Default(c)
		router.Use(sessions.Sessions(input.Login, store))
		session.Set("pass", input.Password)
		session.Set("login", input.Login)
		errSave := session.Save()
		if errSave != nil {
			c.JSON(http.StatusBadRequest,
				gin.H{"error": errSave.Error()})
			return
		}
		c.JSON(200, gin.H{"Status": "Succeed"})
	}
}

func GinCheck(c *gin.Context) {
	var input *Transport.AuthStruct
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": err.Error()})
		return
	} else {
		b := check(input, c)
		c.JSON(200, gin.H{"Status": b})
	}
}

func GinAcceptLoadUser(c *gin.Context) {
	var input *Transport.AcceptLoadUserHelpRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": err.Error()})
		return
	} else {
		b := check(input.Auth, c)
		if b {
			user, errGrpcClient := grpcClient.AcceptLoadUser(context.Background(), &AcceptLoadUserRequest{
				PublicKey:  input.PublicKey,
				PrivateKey: input.PrivateKey,
			})
			fmt.Println(errGrpcClient)
			if errGrpcClient != nil {
				c.JSON(500,
					gin.H{"error": errGrpcClient.Error()})
				return
			}
			c.JSON(200, gin.H{"AcceptLoadUserResponse": user})
		} else {
			c.JSON(403, gin.H{"Status": "Access Denied"})
		}

	}
}

func GinAcceptNewUser(c *gin.Context) {
	var input *Transport.AcceptNewUserHelpRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": err.Error()})
		return
	} else {
		b := check(input.Auth, c)
		if b {
			privateKey, errGrpcClient := grpcClient.AcceptNewUser(context.Background(), &AcceptNewUserRequest{
				Pass:      input.Pass,
				Salt:      input.Salt,
				PublicKey: input.PublicKey,
			})
			fmt.Println(errGrpcClient)
			if errGrpcClient != nil {
				c.JSON(500,
					gin.H{"error": errGrpcClient.Error()})
				return
			}
			c.JSON(200, gin.H{"AcceptNewUserHelpResponse": privateKey})
		} else {
			c.JSON(403, gin.H{"Status": "Access Denied"})
		}

	}
}

func GinVote(c *gin.Context) {
	var input *Transport.VoteHelpRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": err.Error()})
		return
	} else {
		b := check(input.Auth, c)
		if b {
			status, errGrpcClient := grpcClient.Vote(context.Background(), &VoteRequest{
				Receiver: input.Receiver,
				Sender:   input.Sender,
				Master:   input.Master,
				Num:      input.Num,
			})
			fmt.Println(errGrpcClient)
			if errGrpcClient != nil {
				c.JSON(500,
					gin.H{"error": errGrpcClient.Error()})
				return
			}
			c.JSON(200, gin.H{"Vote": status})
		} else {
			c.JSON(403, gin.H{"Status": "Access Denied"})
		}

	}
}

func GinSoloWinner(c *gin.Context) {
	var input *Transport.SoloWinnerHelpRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": err.Error()})
		return
	} else {
		b := check(input.Auth, c)
		if b {
			object, errGrpcClient := grpcClient.SoloWinner(context.Background(), &SoloWinnerRequest{
				Master: input.Master,
			})
			fmt.Println(errGrpcClient)
			if errGrpcClient != nil {
				c.JSON(500,
					gin.H{"error": errGrpcClient.Error()})
				return
			}
			c.JSON(200, gin.H{"SoloWinnerObject": object})
		} else {
			c.JSON(403, gin.H{"Status": "Access Denied"})
		}

	}
}

func GinWinnersList(c *gin.Context) {
	var input *Transport.WinnersListHelpRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": err.Error()})
		return
	} else {
		b := check(input.Auth, c)
		if b {
			list, errGrpcClient := grpcClient.WinnersList(context.Background(), &WinnersListRequest{
				Master: input.Master,
			})
			fmt.Println(errGrpcClient)
			if errGrpcClient != nil {
				c.JSON(500,
					gin.H{"error": errGrpcClient.Error()})
				return
			}
			c.JSON(200, gin.H{"WinnersList": list})
		} else {
			c.JSON(403, gin.H{"Status": "Access Denied"})
		}

	}
}

func GinViewCandidates(c *gin.Context) {
	var input *Transport.CallViewCandidatesHelpRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": err.Error()})
		return
	} else {
		b := check(input.Auth, c)
		if b {
			list, errGrpcClient := grpcClient.CallViewCandidates(context.Background(), &CallViewCandidatesRequest{
				Master: input.Master,
			})
			fmt.Println(errGrpcClient)
			if errGrpcClient != nil {
				c.JSON(500,
					gin.H{"error": errGrpcClient.Error()})
				return
			}
			c.JSON(200, gin.H{"CandidatesList": list})
		} else {
			c.JSON(403, gin.H{"Status": "Access Denied"})
		}

	}
}

func GinNewCandidates(c *gin.Context) {
	var input *Transport.CallNewCandidateHelpRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": err.Error()})
		return
	} else {
		b := check(input.Auth, c)
		if b {
			candidate, errGrpcClient := grpcClient.CallNewCandidate(context.Background(), &CallNewCandidateRequest{
				Description: input.Description,
				Affiliation: input.Affiliation,
			})
			fmt.Println(errGrpcClient)
			if errGrpcClient != nil {
				c.JSON(500,
					gin.H{"error": errGrpcClient.Error()})
				return
			}
			c.JSON(200, gin.H{"electionsObject": candidate})
		} else {
			c.JSON(403, gin.H{"Status": "Access Denied"})
		}

	}
}

func GinCallCreateVoters(c *gin.Context) {
	var input *Transport.CallCreateVotersHelpRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": err.Error()})
		return
	} else {
		b := check(input.Auth, c)
		if b {
			voter, errGrpcClient := grpcClient.CallCreateVoters(context.Background(), &CallCreateVotersRequest{
				Voter:  input.Voter,
				Master: input.Master,
			})
			fmt.Println(errGrpcClient)
			if errGrpcClient != nil {
				c.JSON(500,
					gin.H{"error": errGrpcClient.Error()})
				return
			}
			c.JSON(200, gin.H{"voterObjects": voter})
		} else {
			c.JSON(403, gin.H{"Status": "Access Denied"})
		}

	}
}

func GinGetFull(c *gin.Context) {
	var input *Transport.AuthStruct
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": err.Error()})
		return
	} else {
		b := check(input, c)
		if b {
			chain, errGrpcClient := grpcClient.GetFullChain(context.Background(), &googleProtobuf.Empty{})
			fmt.Println(errGrpcClient)
			if errGrpcClient != nil {
				c.JSON(500,
					gin.H{"error": errGrpcClient.Error()})
				return
			}
			c.JSON(200, gin.H{"fullChain": chain})
		} else {
			c.JSON(403, gin.H{"Status": "Access Denied"})
		}

	}
}

func GinGetPartOfChain(c *gin.Context) {
	var input *Transport.PartOfChainRequestHelpRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": err.Error()})
		return
	} else {
		b := check(input.Auth, c)
		if b {
			chain, errGrpcClient := grpcClient.GetPartOfChain(context.Background(), &GetPartOfChainRequest{
				Master: input.Master,
			})
			fmt.Println(errGrpcClient)
			if errGrpcClient != nil {
				c.JSON(500,
					gin.H{"error": errGrpcClient.Error()})
				return
			}
			c.JSON(200, gin.H{"partOfChain": chain})
		} else {
			c.JSON(403, gin.H{"Status": "Access Denied"})
		}

	}
}

func GinGetChainSize(c *gin.Context) {
	var input *Transport.ChainSizeHelpRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": err.Error()})
		return
	} else {
		b := check(input.Auth, c)
		if b {
			chain, errGrpcClient := grpcClient.GetPartOfChain(context.Background(), &GetPartOfChainRequest{
				Master: input.Master,
			})
			fmt.Println(errGrpcClient)
			if errGrpcClient != nil {
				c.JSON(500,
					gin.H{"error": errGrpcClient.Error()})
				return
			}
			c.JSON(200, gin.H{"partOfChain": chain})
		} else {
			c.JSON(403, gin.H{"Status": "Access Denied"})
		}

	}
}

func GinNewChain(c *gin.Context) {
	var input *Transport.NewChainHelpRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": err.Error()})
		return
	} else {
		b := check(input.Auth, c)
		if b {
			chain, errGrpcClient := grpcClient.NewChain(context.Background(), &NewChainRequest{
				Master:     input.Master,
				VotesCount: input.Count,
				LimitTime:  input.LimitTime,
			})
			fmt.Println(errGrpcClient)
			if errGrpcClient != nil {
				c.JSON(500,
					gin.H{"error": errGrpcClient.Error()})
				return
			}
			c.JSON(200, gin.H{"partOfChain": chain})
		} else {
			c.JSON(403, gin.H{"Status": "Access Denied"})
		}

	}
}

func check(input *Transport.AuthStruct, c *gin.Context) bool {
	fmt.Println(input)
	router.Use(sessions.Sessions(input.Login, store))
	session := sessions.Default(c)
	if session.Get("pass") == input.Password && session.Get("login") == input.Login {
		return true
	} else {
		return false
	}
}
