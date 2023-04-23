package client

import (
	"context"
	"errors"

	sdkmath "cosmossdk.io/math"
	"github.com/bnb-chain/greenfield-go-sdk/types"
	gnfdTypes "github.com/bnb-chain/greenfield/types"
	permTypes "github.com/bnb-chain/greenfield/x/permission/types"
	storageTypes "github.com/bnb-chain/greenfield/x/storage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Group interface {
	// CreateGroup create a new group on greenfield chain the group members can be initialized  or not
	CreateGroup(ctx context.Context, groupName string, opt types.CreateGroupOptions) (string, error)
	// DeleteGroup send DeleteGroup txn to greenfield chain and return txn hash
	DeleteGroup(ctx context.Context, groupName string, opt types.DeleteGroupOption) (string, error)
	// UpdateGroupMember support adding or removing members from the group and return the txn hash
	// groupOwnerAddr indicates the HEX-encoded string of the group owner address
	// addAddresses indicates the HEX-encoded string list of the member addresses to be added
	// removeAddresses indicates the HEX-encoded string list of the member addresses to be removed
	UpdateGroupMember(ctx context.Context, groupName string, groupOwnerAddr string,
		addAddresses, removeAddresses []string, opts types.UpdateGroupMemberOption) (string, error)
	// LeaveGroup make the member leave the specific group
	// groupOwnerAddr indicates the HEX-encoded string of the group owner address
	LeaveGroup(ctx context.Context, groupName string, groupOwnerAddr string, opt types.LeaveGroupOption) (string, error)
	// HeadGroup query the groupInfo on chain, return the group info if exists return err info if group not exist
	// groupOwnerAddr indicates the HEX-encoded string of the group owner address
	HeadGroup(ctx context.Context, groupName string, groupOwnerAddr string) (*storageTypes.GroupInfo, error)
	// HeadGroupMember query the group member info on chain, return true if the member exists in group
	// groupOwnerAddr indicates the HEX-encoded string of the group owner address
	// headMember indicates the HEX-encoded string of the group member address
	HeadGroupMember(ctx context.Context, groupName string, groupOwner, headMember string) bool
	// PutGroupPolicy apply group policy to user specified by principalAddr, the sender need to be the owner of the group
	// principalAddr indicates the HEX-encoded string of the principal address
	PutGroupPolicy(ctx context.Context, groupName string, principalAddr string, statements []*permTypes.Statement, opt types.PutPolicyOption) (string, error)
	// DeleteGroupPolicy  delete group policy of the principal, the sender need to be the owner of the group
	// principalAddr indicates the HEX-encoded string of the principal address
	DeleteGroupPolicy(ctx context.Context, groupName string, principalAddr string, opt types.DeletePolicyOption) (string, error)
	// GetBucketPolicyOfGroup get the bucket policy info of the group specified by group id
	// it queries a bucket policy that grants permission to a group
	GetBucketPolicyOfGroup(ctx context.Context, bucketName string, groupId uint64) (*permTypes.Policy, error)
	// GetObjectPolicyOfGroup get the object policy info of the group specified by group id
	// it queries an object policy that grants permission to a group
	GetObjectPolicyOfGroup(ctx context.Context, bucketName, objectName string, groupId uint64) (*permTypes.Policy, error)
}

// CreateGroup create a new group on greenfield chain, the group members can be initialized or not
func (c *client) CreateGroup(ctx context.Context, groupName string, opt types.CreateGroupOptions) (string, error) {
	createGroupMsg := storageTypes.NewMsgCreateGroup(c.MustGetDefaultAccount().GetAddress(), groupName, opt.InitGroupMember)
	return c.sendTxn(ctx, createGroupMsg, opt.TxOpts)
}

// DeleteGroup send DeleteGroup txn to greenfield chain and return txn hash
func (c *client) DeleteGroup(ctx context.Context, groupName string, opt types.DeleteGroupOption) (string, error) {
	deleteGroupMsg := storageTypes.NewMsgDeleteGroup(c.MustGetDefaultAccount().GetAddress(), groupName)
	return c.sendTxn(ctx, deleteGroupMsg, opt.TxOpts)
}

// UpdateGroupMember support adding or removing members from the group and return the txn hash
func (c *client) UpdateGroupMember(ctx context.Context, groupName string, groupOwnerAddr string,
	addAddresses, removeAddresses []string, opts types.UpdateGroupMemberOption) (string, error) {
	groupOwner, err := sdk.AccAddressFromHexUnsafe(groupOwnerAddr)
	if err != nil {
		return "", err
	}
	if groupName == "" {
		return "", errors.New("group name is empty")
	}

	if len(addAddresses) == 0 && len(removeAddresses) == 0 {
		return "", errors.New("no update member")
	}

	addMembers := make([]sdk.AccAddress, 0)
	removeMembers := make([]sdk.AccAddress, 0)

	for _, addr := range addAddresses {
		member, err := sdk.AccAddressFromHexUnsafe(addr)
		if err != nil {
			return "", err
		}
		addMembers = append(addMembers, member)
	}

	for _, addr := range removeAddresses {
		member, err := sdk.AccAddressFromHexUnsafe(addr)
		if err != nil {
			return "", err
		}
		removeMembers = append(removeMembers, member)
	}

	updateGroupMsg := storageTypes.NewMsgUpdateGroupMember(c.MustGetDefaultAccount().GetAddress(), groupOwner, groupName, addMembers, removeMembers)

	return c.sendTxn(ctx, updateGroupMsg, opts.TxOpts)
}

// LeaveGroup make the member leave the specific group
func (c *client) LeaveGroup(ctx context.Context, groupName string, groupOwnerAddr string, opt types.LeaveGroupOption) (string, error) {
	groupOwner, err := sdk.AccAddressFromHexUnsafe(groupOwnerAddr)
	if err != nil {
		return "", err
	}
	leaveGroupMsg := storageTypes.NewMsgLeaveGroup(c.MustGetDefaultAccount().GetAddress(), groupOwner, groupName)
	return c.sendTxn(ctx, leaveGroupMsg, opt.TxOpts)
}

// HeadGroup query the groupInfo on chain, return the group info if exists
// return err info if group not exist
func (c *client) HeadGroup(ctx context.Context, groupName string, groupOwnerAddr string) (*storageTypes.GroupInfo, error) {
	headGroupRequest := storageTypes.QueryHeadGroupRequest{
		GroupOwner: groupOwnerAddr,
		GroupName:  groupName,
	}

	headGroupResponse, err := c.chainClient.HeadGroup(ctx, &headGroupRequest)
	if err != nil {
		return nil, err
	}

	return headGroupResponse.GroupInfo, nil
}

// HeadGroupMember query the group member info on chain, return true if the member exists in group
func (c *client) HeadGroupMember(ctx context.Context, groupName string, groupOwnerAddr, headMemberAddr string) bool {
	headGroupRequest := storageTypes.QueryHeadGroupMemberRequest{
		GroupName:  groupName,
		GroupOwner: groupOwnerAddr,
		Member:     headMemberAddr,
	}

	_, err := c.chainClient.HeadGroupMember(ctx, &headGroupRequest)
	return err == nil
}

// PutGroupPolicy apply group policy to user specified by principalAddr, the sender need to be the owner of the group
func (c *client) PutGroupPolicy(ctx context.Context, groupName string, principalAddr string,
	statements []*permTypes.Statement, opt types.PutPolicyOption) (string, error) {
	sender := c.MustGetDefaultAccount().GetAddress()

	resource := gnfdTypes.NewGroupGRN(sender, groupName)

	principal, err := sdk.AccAddressFromHexUnsafe(principalAddr)
	if err != nil {
		return "", err
	}

	putPolicyMsg := storageTypes.NewMsgPutPolicy(sender, resource.String(),
		permTypes.NewPrincipalWithAccount(principal), statements, opt.PolicyExpireTime)

	return c.sendPutPolicyTxn(ctx, putPolicyMsg, opt.TxOpts)
}

// GetBucketPolicyOfGroup get the bucket policy info of the group specified by group id
// it queries a bucket policy that grants permission to a group
func (c *client) GetBucketPolicyOfGroup(ctx context.Context, bucketName string, groupId uint64) (*permTypes.Policy, error) {
	resource := gnfdTypes.NewBucketGRN(bucketName).String()

	queryPolicy := storageTypes.QueryPolicyForGroupRequest{
		Resource:         resource,
		PrincipalGroupId: sdkmath.NewUint(groupId).String(),
	}

	queryPolicyResp, err := c.chainClient.QueryPolicyForGroup(ctx, &queryPolicy)
	if err != nil {
		return nil, err
	}

	return queryPolicyResp.Policy, nil
}

// GetObjectPolicyOfGroup get the object policy info of the group specified by group id
// it queries an object policy that grants permission to a group
func (c *client) GetObjectPolicyOfGroup(ctx context.Context, bucketName, objectName string, groupId uint64) (*permTypes.Policy, error) {
	resource := gnfdTypes.NewObjectGRN(bucketName, objectName)
	queryPolicy := storageTypes.QueryPolicyForGroupRequest{
		Resource:         resource.String(),
		PrincipalGroupId: sdkmath.NewUint(groupId).String(),
	}

	queryPolicyResp, err := c.chainClient.QueryPolicyForGroup(ctx, &queryPolicy)
	if err != nil {
		return nil, err
	}

	return queryPolicyResp.Policy, nil
}

// DeleteGroupPolicy delete group policy of the principal, the sender need to be the owner of the group
func (c *client) DeleteGroupPolicy(ctx context.Context, groupName string, principalAddr string, opt types.DeletePolicyOption) (string, error) {
	sender := c.MustGetDefaultAccount().GetAddress()
	resource := gnfdTypes.NewGroupGRN(sender, groupName).String()

	addr, err := sdk.AccAddressFromHexUnsafe(principalAddr)
	if err != nil {
		return "", err
	}

	principal := permTypes.NewPrincipalWithAccount(addr)

	return c.sendDelPolicyTxn(ctx, sender, resource, principal, opt.TxOpts)
}