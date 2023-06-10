package Transport

import "google.golang.org/protobuf/types/known/timestamppb"

type AuthStruct struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
type NewChainHelpRequest struct {
	Master    string                 `form:"master" json:"master"`
	Count     int32                  `form:"count" json:"count"`
	LimitTime *timestamppb.Timestamp `form:"limitTime" json:"limitTime"`
	Auth      *AuthStruct            `form:"auth" json:"auth"`
}
type NewChainHelpResponse struct {
	Status string      `form:"status" json:"status"`
	Auth   *AuthStruct `form:"auth" json:"auth"`
}
type CallCreateVotersHelpRequest struct {
	Voter  string      `form:"voter" json:"voter"`
	Master string      `form:"master" json:"master"`
	Auth   *AuthStruct `form:"auth" json:"auth"`
}
type CallCreateVotersHelpResponse struct {
	User BlockchainUser `form:"user" json:"User"`
	Auth *AuthStruct    `form:"auth" json:"auth"`
}
type BlockchainUser struct {
	Id          string `form:"id" json:"Id"`
	PublicKey   string `form:"publicKey" json:"PublicKey"`
	IsUsed      bool   `form:"isUsed" json:"IsUsed"`
	Affiliation string `form:"affiliation" json:"Affiliation"`
}
type CallViewCandidatesHelpRequest struct {
	Master string      `form:"master" json:"master"`
	Auth   *AuthStruct `form:"auth" json:"auth"`
}
type CallViewCandidatesHelpResponse struct {
	ElectionSubjects BlockchainElectionSubjects `form:"electionSubjects" json:"ElectionSubjects"`
	Auth             *AuthStruct                `form:"auth" json:"auth"`
}
type BlockchainElectionSubjects struct {
	Id                string `form:"id" json:"Id"`
	PublicKey         string `form:"publicKey" json:"PublicKey"`
	Description       bool   `form:"description" json:"Description"`
	VotingAffiliation string `form:"votingAffiliation" json:"VotingAffiliation"`
}
type CallNewCandidateHelpRequest struct {
	Description string      `form:"description" json:"description"`
	Affiliation string      `form:"affiliation" json:"affiliation"`
	Auth        *AuthStruct `form:"auth" json:"auth"`
}
type CallNewCandidateHelpResponse struct {
	ElectionSubjects BlockchainElectionSubjects `form:"electionSubjects" json:"electionSubjects"`
	Auth             *AuthStruct                `form:"auth" json:"auth"`
}
type AcceptLoadUserHelpRequest struct {
	PublicKey  string      `form:"publicKey" json:"publicKey"`
	PrivateKey string      `form:"privateKey" json:"privateKey"`
	Auth       *AuthStruct `form:"auth" json:"auth"`
}
type AcceptLoadUserHelpResponse struct {
	User BlockchainUser `form:"user" json:"User"`
	Auth *AuthStruct    `form:"auth" json:"auth"`
}
type AcceptNewUserHelpRequest struct {
	Pass      string      `form:"pass" json:"pass"`
	PublicKey string      `form:"publicKey" json:"publicKey"`
	Salt      string      `form:"salt" json:"salt"`
	Auth      *AuthStruct `form:"auth" json:"auth"`
}
type AcceptNewUserHelpResponse struct {
	PrivateKey string      `form:"privateKey" json:"privateKey"`
	Auth       *AuthStruct `form:"auth" json:"auth"`
}
type VoteHelpRequest struct {
	Receiver string      `form:"receiver" json:"receiver"`
	Sender   string      `form:"sender" json:"sender"`
	Master   string      `form:"master" json:"master"`
	Num      int64       `form:"num" json:"num"`
	Auth     *AuthStruct `form:"auth" json:"auth"`
}
type SoloWinnerHelpRequest struct {
	Master string      `form:"master" json:"master"`
	Auth   *AuthStruct `form:"auth" json:"auth"`
}
type SoloWinnerHelpResponse struct {
	SoloWinnerObject *ContractElectionsList `form:"soloWinnerObject" json:"soloWinnerObject"`
	Auth             *AuthStruct            `form:"auth" json:"auth"`
}
type ContractElectionsList struct {
	ElectionList *BlockchainElectionSubjects `form:"electionList" json:"electionList"`
	Balance      string                      `form:"balance" json:"balance"`
}
type WinnersListHelpRequest struct {
	Master string      `form:"master" json:"master"`
	Auth   *AuthStruct `form:"auth" json:"auth"`
}
type WinnersListHelpResponse struct {
	SoloWinnerObject []*BlockchainElectionSubjects `form:"soloWinnerObject" json:"soloWinnerObject"`
	Auth             *AuthStruct                   `form:"auth" json:"auth"`
}
type PartOfChainRequestHelpRequest struct {
	Master string      `form:"master" json:"master"`
	Auth   *AuthStruct `form:"auth" json:"auth"`
}
type ChainSizeHelpRequest struct {
	Master string      `form:"master" json:"master"`
	Auth   *AuthStruct `form:"auth" json:"auth"`
}
