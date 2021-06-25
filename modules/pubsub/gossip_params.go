package pubsub

import (
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	pubsub_pb "github.com/libp2p/go-libp2p-pubsub/pb"
	"github.com/minio/blake2b-simd"
)

var topicParams = map[string]*pubsub.TopicScoreParams{
	topic: {
		// expected > 1 tx/second
		TopicWeight: 0.1, // max cap is 5, single invalid message is -100

		// 1 tick per second, maxes at 1 hour
		TimeInMeshWeight:  0.0002778, // ~1/3600
		TimeInMeshQuantum: time.Second,
		TimeInMeshCap:     1,

		// deliveries decay after 10min, cap at 100 tx
		FirstMessageDeliveriesWeight: 0.5, // max value is 50
		FirstMessageDeliveriesDecay:  pubsub.ScoreParameterDecay(10 * time.Minute),
		FirstMessageDeliveriesCap:    100, // 100 messages in 10 minutes

		// Mesh Delivery Failure is currently turned off for messages
		// This is on purpose as the network is still too small, which results in
		// asymmetries and potential unmeshing from negative scores.
		// // tracks deliveries in the last minute
		// // penalty activates at 1 min and expects 2.5 txs
		// MeshMessageDeliveriesWeight:     -16, // max penalty is -100
		// MeshMessageDeliveriesDecay:      pubsub.ScoreParameterDecay(time.Minute),
		// MeshMessageDeliveriesCap:        100, // 100 txs in a minute
		// MeshMessageDeliveriesThreshold:  2.5, // 60/12/2 txs/minute
		// MeshMessageDeliveriesWindow:     10 * time.Millisecond,
		// MeshMessageDeliveriesActivation: time.Minute,

		// // decays after 5min
		// MeshFailurePenaltyWeight: -16,
		// MeshFailurePenaltyDecay:  pubsub.ScoreParameterDecay(5 * time.Minute),

		// invalid messages decay after 1 hour
		InvalidMessageDeliveriesWeight: -1000,
		InvalidMessageDeliveriesDecay:  pubsub.ScoreParameterDecay(time.Hour),
	},
}

var pgTopicWeights = map[string]float64{
	topic: 1000,
}

func HashMsgId(m *pubsub_pb.Message) string {
	hash := blake2b.Sum256(m.Data)
	return string(hash[:])
}

var options = []pubsub.Option{
	// Gossipsubv1.1 configuration
	pubsub.WithFloodPublish(true),
	// pubsub.WithMessageIdFn(HashMsgId),
	pubsub.WithPeerScore(
		&pubsub.PeerScoreParams{
			AppSpecificScore: func(p peer.ID) float64 {
				// return a heavy positive score for bootstrappers so that we don't unilaterally prune
				// them and accept PX from them.
				// we don't do that in the bootstrappers themselves to avoid creating a closed mesh
				// between them (however we might want to consider doing just that)

				// TODO: we want to  plug the application specific score to the node itself in order
				//       to provide feedback to the pubsub system based on observed behaviour
				return 1000
			},
			AppSpecificWeight: 1000,

			// This sets the IP colocation threshold to 5 peers before we apply penalties
			IPColocationFactorThreshold: 5,
			IPColocationFactorWeight:    -100,

			// P7: behavioural penalties, decay after 1hr
			BehaviourPenaltyThreshold: 6,
			BehaviourPenaltyWeight:    -10,
			BehaviourPenaltyDecay:     pubsub.ScoreParameterDecay(time.Hour),

			DecayInterval: pubsub.DefaultDecayInterval,
			DecayToZero:   pubsub.DefaultDecayToZero,

			// this retains non-positive scores for 6 hours
			RetainScore: 6 * time.Hour,

			// topic parameters
			Topics: topicParams,
		},
		&pubsub.PeerScoreThresholds{
			GossipThreshold:             -500,
			PublishThreshold:            -1000,
			GraylistThreshold:           -2500,
			AcceptPXThreshold:           1000,
			OpportunisticGraftThreshold: 3.5,
		},
	),
}

// enable Peer eXchange on bootstrappers
// if isBootstrapNode {
// 	// turn off the mesh in bootstrappers -- only do gossip and PX
// 	pubsub.GossipSubD = 0
// 	pubsub.GossipSubDscore = 0
// 	pubsub.GossipSubDlo = 0
// 	pubsub.GossipSubDhi = 0
// 	pubsub.GossipSubDout = 0
// 	pubsub.GossipSubDlazy = 64
// 	pubsub.GossipSubGossipFactor = 0.25
// 	pubsub.GossipSubPruneBackoff = 5 * time.Minute
// 	// turn on PX
// 	options = append(options, pubsub.WithPeerExchange(true))
// }

func init() {
	var pgParams *pubsub.PeerGaterParams = pubsub.NewPeerGaterParams(
		0.33,
		pubsub.ScoreParameterDecay(2*time.Minute),
		pubsub.ScoreParameterDecay(time.Hour),
	).WithTopicDeliveryWeights(pgTopicWeights)

	options = append(options, pubsub.WithPeerGater(pgParams))

	allowTopics := []string{
		topic,
	}
	options = append(options,
		pubsub.WithSubscriptionFilter(
			pubsub.WrapLimitSubscriptionFilter(
				pubsub.NewAllowlistSubscriptionFilter(allowTopics...),
				100)))

	// still instantiate a tracer for collecting metrics
}
