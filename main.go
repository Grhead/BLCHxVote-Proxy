package main

import (
	"Vox2-Proxy/Transport"
	. "Vox2-Proxy/Transport/PBs"
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	intersessions "github.com/gin-contrib/sessions/gorm"
	"github.com/gin-gonic/gin"
	googleProtobuf "github.com/golang/protobuf/ptypes/empty"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"net/http"
	"strings"
)

type Store interface {
	sessions.Store
}
type TableAuth struct {
	Login    string
	Password string
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

	store = intersessions.NewStore(db, false, []byte("secret"))
	router = gin.Default()
	router.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"PUT", "PATCH", "POST", "GET", "DELETE"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type", "Accept-Encoding"},
		ExposeHeaders:    []string{"Content-Length", "Access-Control-Allow-Origin", "Access-Control-Allow-Credentials", "Access-Control-Allow-Headers", "Access-Control-Allow-Methods"},
		AllowCredentials: true,
	}))
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
			gin.H{"error": err})
		return
	} else {
		db, errGorm := gorm.Open(sqlite.Open("Database/AuthDB.db"), &gorm.Config{})
		if errGorm != nil {
			c.JSON(http.StatusBadRequest,
				gin.H{"error": errGorm})
			return
		}
		session := sessions.Default(c)
		var AuthArray []*TableAuth
		db.Raw("SELECT Login, Password FROM AuthDataTable").Scan(&AuthArray)
		for _, v := range AuthArray {
			fmt.Println(v)
		}
		if input.Password != "" && input.Login != "" && len(input.Password) >= 8 && len(input.Login) >= 8 && session.Get(input.Password) != "" {
			db.Exec("INSERT INTO AuthDataTable VALUES ($1, $2, $3)", uuid.New().String(), HashSum(input.Login), HashSum(input.Password))
			router.Use(sessions.Sessions(input.Login, store))
			session.Set("pass", input.Password)
			session.Set("login", input.Login)
			errSave := session.Save()
			if errSave != nil {
				c.JSON(http.StatusBadRequest,
					gin.H{"error": errSave})
				return
			}
			c.JSON(200, gin.H{"status": "Succeed"})
		} else {
			c.JSON(200, gin.H{"status": "Auth Error"})
		}

	}
}

func GinCheck(c *gin.Context) {
	var input *Transport.AuthStruct
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": err})
		return
	} else {
		b := check(input, c)
		c.JSON(200, gin.H{"status": b})
	}
}

func GinAcceptLoadUser(c *gin.Context) {
	var input *Transport.AcceptLoadUserHelpRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": err})
		return
	} else {
		b := check(input.Auth, c)
		if b {
			user, errGrpcClient := grpcClient.AcceptLoadUser(context.Background(), &AcceptLoadUserRequest{
				PublicKey:  strings.ToLower(input.PublicKey),
				PrivateKey: strings.ToLower(input.PrivateKey),
			})
			fmt.Println(errGrpcClient)
			if errGrpcClient != nil {
				c.JSON(500,
					gin.H{"error": errGrpcClient})
				return
			}
			c.JSON(200, gin.H{"acceptLoadUserResponse": user.User})
		} else {
			c.JSON(403, gin.H{"status": "Access Denied"})
		}
	}
}

func GinAcceptNewUser(c *gin.Context) {
	var input *Transport.AcceptNewUserHelpRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": err})
		return
	} else {
		b := check(input.Auth, c)
		if b {
			privateKey, errGrpcClient := grpcClient.AcceptNewUser(context.Background(), &AcceptNewUserRequest{
				Pass:      strings.ToLower(input.Pass),
				Salt:      strings.ToLower(input.Salt),
				PublicKey: strings.ToLower(input.PublicKey),
			})
			fmt.Println(errGrpcClient)
			if errGrpcClient != nil {
				c.JSON(500,
					gin.H{"error": errGrpcClient})
				return
			}
			c.JSON(200, gin.H{"acceptNewUserHelpResponse": privateKey})
		} else {
			c.JSON(403, gin.H{"status": "Access Denied"})
		}

	}
}

func GinVote(c *gin.Context) {
	var input *Transport.VoteHelpRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": err})
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
					gin.H{"error": errGrpcClient})
				return
			}
			c.JSON(200, gin.H{"vote": status})
		} else {
			c.JSON(403, gin.H{"status": "Access Denied"})
		}

	}
}

func GinSoloWinner(c *gin.Context) {
	var input *Transport.SoloWinnerHelpRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": err})
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
					gin.H{"error": errGrpcClient})
				return
			}
			c.JSON(200, gin.H{"soloWinnerObject": object})
		} else {
			c.JSON(403, gin.H{"status": "Access Denied"})
		}

	}
}

func GinWinnersList(c *gin.Context) {
	var input *Transport.WinnersListHelpRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": err})
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
					gin.H{"error": errGrpcClient})
				return
			}
			c.JSON(200, gin.H{"winnersList": list})
		} else {
			c.JSON(403, gin.H{"status": "Access Denied"})
		}

	}
}

func GinViewCandidates(c *gin.Context) {
	var input *Transport.CallViewCandidatesHelpRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": err})
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
					gin.H{"error": errGrpcClient})
				return
			}
			c.JSON(200, gin.H{"candidatesList": list})
		} else {
			c.JSON(403, gin.H{"status": "Access Denied"})
		}

	}
}

func GinNewCandidates(c *gin.Context) {
	var input *Transport.CallNewCandidateHelpRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": err})
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
					gin.H{"error": errGrpcClient})
				return
			}
			c.JSON(200, gin.H{" else ": candidate})
		} else {
			c.JSON(403, gin.H{"status": "Access Denied"})
		}

	}
}

func GinCallCreateVoters(c *gin.Context) {
	var input *Transport.CallCreateVotersHelpRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": err})
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
					gin.H{"error": errGrpcClient})
				return
			}
			c.JSON(200, gin.H{"voterObjects": voter.User, "identities": voter.Identifier})
		} else {
			c.JSON(403, gin.H{"status": "Access Denied"})
		}

	}
}

func GinGetFull(c *gin.Context) {
	var input *Transport.AuthStruct
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": err})
		return
	} else {
		b := check(input, c)
		if b {
			chain, errGrpcClient := grpcClient.GetFullChain(context.Background(), &googleProtobuf.Empty{})
			fmt.Println(errGrpcClient)
			if errGrpcClient != nil {
				c.JSON(500,
					gin.H{"error": errGrpcClient})
				return
			}
			c.JSON(200, gin.H{"fullChain": chain})
		} else {
			c.JSON(403, gin.H{"status": "Access Denied"})
		}

	}
}

func GinGetPartOfChain(c *gin.Context) {
	var input *Transport.PartOfChainRequestHelpRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": err})
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
					gin.H{"error": errGrpcClient})
				return
			}
			c.JSON(200, gin.H{"partOfChain": chain})
		} else {
			c.JSON(403, gin.H{"status": "Access Denied"})
		}

	}
}

func GinGetChainSize(c *gin.Context) {
	var input *Transport.ChainSizeHelpRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": err})
		return
	} else {
		b := check(input.Auth, c)
		if b {
			chain, errGrpcClient := grpcClient.ChainSize(context.Background(), &ChainSizeRequest{
				Master: input.Master,
			})
			fmt.Println(errGrpcClient)
			if errGrpcClient != nil {
				c.JSON(500,
					gin.H{"error": errGrpcClient})
				return
			}
			c.JSON(200, gin.H{"partOfChain": chain})
		} else {
			c.JSON(403, gin.H{"status": "Access Denied"})
		}

	}
}

func GinNewChain(c *gin.Context) {
	var input *Transport.NewChainHelpRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": err})
		return
	} else {
		b := check(input.Auth, c)
		if b {
			chain, errGrpcClient := grpcClient.NewChain(context.Background(), &NewChainRequest{
				Master:     input.Master,
				VotesCount: input.Count,
				LimitTime:  input.LimitTime,
			})
			log.Println("::", errGrpcClient)
			if errGrpcClient != nil {
				c.JSON(200,
					gin.H{"error": errGrpcClient})
				return
			}
			log.Println("::", "OOOROR")
			c.JSON(200, gin.H{"partOfChain": chain})
		} else {
			c.JSON(403, gin.H{"status": "Access Denied"})
		}

	}
}

func check(input *Transport.AuthStruct, c *gin.Context) bool {
	fmt.Println(input)
	db, errGorm := gorm.Open(sqlite.Open("Database/AuthDB.db"), &gorm.Config{})
	if errGorm != nil {
		return false
	}
	var AuthArray *TableAuth
	db.Raw("SELECT Login, Password FROM AuthDataTable WHERE $1 = Login AND $2 = Password",
		HashSum(input.Login),
		HashSum(input.Password)).
		Scan(&AuthArray)
	//router.Use(sessions.Sessions(input.Login, store))
	//session := sessions.Default(c)
	if AuthArray != nil {
		return true
	} else {
		return false
	}
	//if session.Get("pass") == input.Password && session.Get("login") == input.Login {
	//	return true
	//} else {
	//	return false
	//}
}
func HashSum(data string) string {
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash[:])
}
