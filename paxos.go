package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// ProposalID 提案ID，包含轮次号和节点ID
type ProposalID struct {
	Round  int64
	NodeID int
}

// Less 比较两个提案ID的大小
func (pid ProposalID) Less(other ProposalID) bool {
	if pid.Round != other.Round {
		return pid.Round < other.Round
	}
	return pid.NodeID < other.NodeID
}

// Equals 判断两个提案ID是否相等
func (pid ProposalID) Equals(other ProposalID) bool {
	return pid.Round == other.Round && pid.NodeID == other.NodeID
}

// PrepareRequest 准备阶段请求
type PrepareRequest struct {
	ProposalID ProposalID
}

// PrepareResponse 准备阶段响应
type PrepareResponse struct {
	Accepted bool
	// 如果之前接受过提案，返回已接受的提案信息
	AcceptedID    ProposalID
	AcceptedValue interface{}
}

// AcceptRequest 接受阶段请求
type AcceptRequest struct {
	ProposalID ProposalID
	Value      interface{}
}

// AcceptResponse 接受阶段响应
type AcceptResponse struct {
	Accepted bool
}

// LearnRequest 学习阶段请求
type LearnRequest struct {
	ProposalID ProposalID
	Value      interface{}
}

// Acceptor 接受者节点
type Acceptor struct {
	nodeID        int
	promisedID    ProposalID
	acceptedID    ProposalID
	acceptedValue interface{}
	mu            sync.RWMutex
}

// NewAcceptor 创建新的接受者
func NewAcceptor(nodeID int) *Acceptor {
	return &Acceptor{
		nodeID: nodeID,
	}
}

// Prepare 处理准备阶段请求
func (a *Acceptor) Prepare(req PrepareRequest) PrepareResponse {
	a.mu.Lock()
	defer a.mu.Unlock()

	// 如果请求的提案ID小于已承诺的提案ID，拒绝
	if req.ProposalID.Less(a.promisedID) {
		return PrepareResponse{Accepted: false}
	}

	// 承诺接受该提案ID
	a.promisedID = req.ProposalID

	// 返回已接受的提案信息（如果有）
	return PrepareResponse{
		Accepted:      true,
		AcceptedID:    a.acceptedID,
		AcceptedValue: a.acceptedValue,
	}
}

// Accept 处理接受阶段请求
func (a *Acceptor) Accept(req AcceptRequest) AcceptResponse {
	a.mu.Lock()
	defer a.mu.Unlock()

	// 如果请求的提案ID小于已承诺的提案ID，拒绝
	if req.ProposalID.Less(a.promisedID) {
		return AcceptResponse{Accepted: false}
	}

	// 接受该提案
	a.promisedID = req.ProposalID
	a.acceptedID = req.ProposalID
	a.acceptedValue = req.Value

	return AcceptResponse{Accepted: true}
}

// Proposer 提案者节点
type Proposer struct {
	nodeID       int
	acceptors    []*Acceptor
	learners     []*Learner
	currentRound int64
	mu           sync.Mutex
}

// NewProposer 创建新的提案者
func NewProposer(nodeID int, acceptors []*Acceptor, learners []*Learner) *Proposer {
	return &Proposer{
		nodeID:    nodeID,
		acceptors: acceptors,
		learners:  learners,
	}
}

// Propose 提出提案
func (p *Proposer) Propose(value interface{}) (bool, interface{}) {
	p.mu.Lock()
	p.currentRound++
	round := p.currentRound
	p.mu.Unlock()

	proposalID := ProposalID{Round: round, NodeID: p.nodeID}

	// 阶段1：准备阶段
	prepareResponses := p.prepare(proposalID)
	if len(prepareResponses) < len(p.acceptors)/2+1 {
		return false, nil // 没有获得多数派支持
	}

	// 检查是否有已接受的值
	var highestAcceptedID ProposalID
	var highestAcceptedValue interface{}
	for _, resp := range prepareResponses {
		if resp.Accepted && resp.AcceptedID.Round > highestAcceptedID.Round {
			highestAcceptedID = resp.AcceptedID
			highestAcceptedValue = resp.AcceptedValue
		}
	}

	// 如果有已接受的值，使用该值；否则使用提议的值
	proposedValue := value
	if highestAcceptedValue != nil {
		proposedValue = highestAcceptedValue
	}

	// 阶段2：接受阶段
	acceptResponses := p.accept(proposalID, proposedValue)
	if len(acceptResponses) < len(p.acceptors)/2+1 {
		return false, nil // 没有获得多数派支持
	}

	// 阶段3：学习阶段
	p.learn(proposalID, proposedValue)

	return true, proposedValue
}

// prepare 准备阶段
func (p *Proposer) prepare(proposalID ProposalID) []PrepareResponse {
	responses := make([]PrepareResponse, 0, len(p.acceptors))

	for _, acceptor := range p.acceptors {
		// 模拟网络延迟
		time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)

		resp := acceptor.Prepare(PrepareRequest{ProposalID: proposalID})
		if resp.Accepted {
			responses = append(responses, resp)
		}
	}

	return responses
}

// accept 接受阶段
func (p *Proposer) accept(proposalID ProposalID, value interface{}) []AcceptResponse {
	responses := make([]AcceptResponse, 0, len(p.acceptors))

	for _, acceptor := range p.acceptors {
		// 模拟网络延迟
		time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)

		resp := acceptor.Accept(AcceptRequest{
			ProposalID: proposalID,
			Value:      value,
		})
		if resp.Accepted {
			responses = append(responses, resp)
		}
	}

	return responses
}

// learn 学习阶段
func (p *Proposer) learn(proposalID ProposalID, value interface{}) {
	for _, learner := range p.learners {
		// 模拟网络延迟
		time.Sleep(time.Duration(rand.Intn(5)) * time.Millisecond)

		learner.Learn(LearnRequest{
			ProposalID: proposalID,
			Value:      value,
		})
	}
}

// Learner 学习者节点
type Learner struct {
	nodeID        int
	learnedValues map[ProposalID]interface{}
	mu            sync.RWMutex
}

// NewLearner 创建新的学习者
func NewLearner(nodeID int) *Learner {
	return &Learner{
		nodeID:        nodeID,
		learnedValues: make(map[ProposalID]interface{}),
	}
}

// Learn 学习已达成共识的值
func (l *Learner) Learn(req LearnRequest) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.learnedValues[req.ProposalID] = req.Value
	fmt.Printf("学习者 %d 学习到提案 %v 的值为: %v\n", l.nodeID, req.ProposalID, req.Value)
}

// GetLearnedValues 获取所有学习到的值
func (l *Learner) GetLearnedValues() map[ProposalID]interface{} {
	l.mu.RLock()
	defer l.mu.RUnlock()

	result := make(map[ProposalID]interface{})
	for k, v := range l.learnedValues {
		result[k] = v
	}
	return result
}

// PaxosCluster Paxos集群
type PaxosCluster struct {
	acceptors []*Acceptor
	proposers []*Proposer
	learners  []*Learner
}

// NewPaxosCluster 创建新的Paxos集群
func NewPaxosCluster(numNodes int) *PaxosCluster {
	acceptors := make([]*Acceptor, numNodes)
	learners := make([]*Learner, numNodes)

	for i := 0; i < numNodes; i++ {
		acceptors[i] = NewAcceptor(i)
		learners[i] = NewLearner(i)
	}

	proposers := make([]*Proposer, numNodes)
	for i := 0; i < numNodes; i++ {
		proposers[i] = NewProposer(i, acceptors, learners)
	}

	return &PaxosCluster{
		acceptors: acceptors,
		proposers: proposers,
		learners:  learners,
	}
}

// Propose 通过指定节点提出提案
func (pc *PaxosCluster) Propose(nodeID int, value interface{}) (bool, interface{}) {
	if nodeID < 0 || nodeID >= len(pc.proposers) {
		return false, nil
	}
	return pc.proposers[nodeID].Propose(value)
}

// GetLearnedValues 获取指定节点学习到的值
func (pc *PaxosCluster) GetLearnedValues(nodeID int) map[ProposalID]interface{} {
	if nodeID < 0 || nodeID >= len(pc.learners) {
		return nil
	}
	return pc.learners[nodeID].GetLearnedValues()
}

// DemoPaxos 演示Paxos算法
func DemoPaxos() {
	fmt.Println("=== Paxos 分布式一致性算法演示 ===")

	// 创建包含5个节点的Paxos集群
	cluster := NewPaxosCluster(5)

	// 模拟多个提案者同时提出提案
	values := []interface{}{"提案A", "提案B", "提案C", "提案D", "提案E"}

	var wg sync.WaitGroup

	for i, value := range values {
		wg.Add(1)
		go func(nodeID int, val interface{}) {
			defer wg.Done()

			fmt.Printf("节点 %d 开始提出提案: %v\n", nodeID, val)
			success, learnedValue := cluster.Propose(nodeID, val)

			if success {
				fmt.Printf("节点 %d 提案成功，学习到的值为: %v\n", nodeID, learnedValue)
			} else {
				fmt.Printf("节点 %d 提案失败\n", nodeID)
			}
		}(i, value)
	}

	wg.Wait()

	fmt.Println("\n=== 各节点学习到的值 ===")
	for i := 0; i < 5; i++ {
		learnedValues := cluster.GetLearnedValues(i)
		fmt.Printf("节点 %d 学习到的值: %v\n", i, learnedValues)
	}
}
