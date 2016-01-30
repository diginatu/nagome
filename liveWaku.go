package nicolive

// LiveWaku is a live broadcast(Waku) of Niconama
type LiveWaku struct {
	Title, BroadID, CommunityID string
	OwnerName, BroadcastToken   string

	StTime, EdTime Time

	UserID, OwnerID string
	UserPremium     bool

	PostKey string

	Addr, Thread string
	Port         int

	OwnerBroad        bool
	OwnerCommentToken string
}
