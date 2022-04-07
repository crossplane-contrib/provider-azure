/*
Copyright 2022 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
// Code generated by angryjet. DO NOT EDIT.

package v1alpha3

import xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"

// GetCondition of this CosmosDBAccount.
func (mg *CosmosDBAccount) GetCondition(ct xpv1.ConditionType) xpv1.Condition {
	return mg.Status.GetCondition(ct)
}

// GetDeletionPolicy of this CosmosDBAccount.
func (mg *CosmosDBAccount) GetDeletionPolicy() xpv1.DeletionPolicy {
	return mg.Spec.DeletionPolicy
}

// GetProviderConfigReference of this CosmosDBAccount.
func (mg *CosmosDBAccount) GetProviderConfigReference() *xpv1.Reference {
	return mg.Spec.ProviderConfigReference
}

/*
GetProviderReference of this CosmosDBAccount.
Deprecated: Use GetProviderConfigReference.
*/
func (mg *CosmosDBAccount) GetProviderReference() *xpv1.Reference {
	return mg.Spec.ProviderReference
}

// GetPublishConnectionDetailsTo of this CosmosDBAccount.
func (mg *CosmosDBAccount) GetPublishConnectionDetailsTo() *xpv1.PublishConnectionDetailsTo {
	return mg.Spec.PublishConnectionDetailsTo
}

// GetWriteConnectionSecretToReference of this CosmosDBAccount.
func (mg *CosmosDBAccount) GetWriteConnectionSecretToReference() *xpv1.SecretReference {
	return mg.Spec.WriteConnectionSecretToReference
}

// SetConditions of this CosmosDBAccount.
func (mg *CosmosDBAccount) SetConditions(c ...xpv1.Condition) {
	mg.Status.SetConditions(c...)
}

// SetDeletionPolicy of this CosmosDBAccount.
func (mg *CosmosDBAccount) SetDeletionPolicy(r xpv1.DeletionPolicy) {
	mg.Spec.DeletionPolicy = r
}

// SetProviderConfigReference of this CosmosDBAccount.
func (mg *CosmosDBAccount) SetProviderConfigReference(r *xpv1.Reference) {
	mg.Spec.ProviderConfigReference = r
}

/*
SetProviderReference of this CosmosDBAccount.
Deprecated: Use SetProviderConfigReference.
*/
func (mg *CosmosDBAccount) SetProviderReference(r *xpv1.Reference) {
	mg.Spec.ProviderReference = r
}

// SetPublishConnectionDetailsTo of this CosmosDBAccount.
func (mg *CosmosDBAccount) SetPublishConnectionDetailsTo(r *xpv1.PublishConnectionDetailsTo) {
	mg.Spec.PublishConnectionDetailsTo = r
}

// SetWriteConnectionSecretToReference of this CosmosDBAccount.
func (mg *CosmosDBAccount) SetWriteConnectionSecretToReference(r *xpv1.SecretReference) {
	mg.Spec.WriteConnectionSecretToReference = r
}

// GetCondition of this MySQLServerFirewallRule.
func (mg *MySQLServerFirewallRule) GetCondition(ct xpv1.ConditionType) xpv1.Condition {
	return mg.Status.GetCondition(ct)
}

// GetDeletionPolicy of this MySQLServerFirewallRule.
func (mg *MySQLServerFirewallRule) GetDeletionPolicy() xpv1.DeletionPolicy {
	return mg.Spec.DeletionPolicy
}

// GetProviderConfigReference of this MySQLServerFirewallRule.
func (mg *MySQLServerFirewallRule) GetProviderConfigReference() *xpv1.Reference {
	return mg.Spec.ProviderConfigReference
}

/*
GetProviderReference of this MySQLServerFirewallRule.
Deprecated: Use GetProviderConfigReference.
*/
func (mg *MySQLServerFirewallRule) GetProviderReference() *xpv1.Reference {
	return mg.Spec.ProviderReference
}

// GetPublishConnectionDetailsTo of this MySQLServerFirewallRule.
func (mg *MySQLServerFirewallRule) GetPublishConnectionDetailsTo() *xpv1.PublishConnectionDetailsTo {
	return mg.Spec.PublishConnectionDetailsTo
}

// GetWriteConnectionSecretToReference of this MySQLServerFirewallRule.
func (mg *MySQLServerFirewallRule) GetWriteConnectionSecretToReference() *xpv1.SecretReference {
	return mg.Spec.WriteConnectionSecretToReference
}

// SetConditions of this MySQLServerFirewallRule.
func (mg *MySQLServerFirewallRule) SetConditions(c ...xpv1.Condition) {
	mg.Status.SetConditions(c...)
}

// SetDeletionPolicy of this MySQLServerFirewallRule.
func (mg *MySQLServerFirewallRule) SetDeletionPolicy(r xpv1.DeletionPolicy) {
	mg.Spec.DeletionPolicy = r
}

// SetProviderConfigReference of this MySQLServerFirewallRule.
func (mg *MySQLServerFirewallRule) SetProviderConfigReference(r *xpv1.Reference) {
	mg.Spec.ProviderConfigReference = r
}

/*
SetProviderReference of this MySQLServerFirewallRule.
Deprecated: Use SetProviderConfigReference.
*/
func (mg *MySQLServerFirewallRule) SetProviderReference(r *xpv1.Reference) {
	mg.Spec.ProviderReference = r
}

// SetPublishConnectionDetailsTo of this MySQLServerFirewallRule.
func (mg *MySQLServerFirewallRule) SetPublishConnectionDetailsTo(r *xpv1.PublishConnectionDetailsTo) {
	mg.Spec.PublishConnectionDetailsTo = r
}

// SetWriteConnectionSecretToReference of this MySQLServerFirewallRule.
func (mg *MySQLServerFirewallRule) SetWriteConnectionSecretToReference(r *xpv1.SecretReference) {
	mg.Spec.WriteConnectionSecretToReference = r
}

// GetCondition of this MySQLServerVirtualNetworkRule.
func (mg *MySQLServerVirtualNetworkRule) GetCondition(ct xpv1.ConditionType) xpv1.Condition {
	return mg.Status.GetCondition(ct)
}

// GetDeletionPolicy of this MySQLServerVirtualNetworkRule.
func (mg *MySQLServerVirtualNetworkRule) GetDeletionPolicy() xpv1.DeletionPolicy {
	return mg.Spec.DeletionPolicy
}

// GetProviderConfigReference of this MySQLServerVirtualNetworkRule.
func (mg *MySQLServerVirtualNetworkRule) GetProviderConfigReference() *xpv1.Reference {
	return mg.Spec.ProviderConfigReference
}

/*
GetProviderReference of this MySQLServerVirtualNetworkRule.
Deprecated: Use GetProviderConfigReference.
*/
func (mg *MySQLServerVirtualNetworkRule) GetProviderReference() *xpv1.Reference {
	return mg.Spec.ProviderReference
}

// GetPublishConnectionDetailsTo of this MySQLServerVirtualNetworkRule.
func (mg *MySQLServerVirtualNetworkRule) GetPublishConnectionDetailsTo() *xpv1.PublishConnectionDetailsTo {
	return mg.Spec.PublishConnectionDetailsTo
}

// GetWriteConnectionSecretToReference of this MySQLServerVirtualNetworkRule.
func (mg *MySQLServerVirtualNetworkRule) GetWriteConnectionSecretToReference() *xpv1.SecretReference {
	return mg.Spec.WriteConnectionSecretToReference
}

// SetConditions of this MySQLServerVirtualNetworkRule.
func (mg *MySQLServerVirtualNetworkRule) SetConditions(c ...xpv1.Condition) {
	mg.Status.SetConditions(c...)
}

// SetDeletionPolicy of this MySQLServerVirtualNetworkRule.
func (mg *MySQLServerVirtualNetworkRule) SetDeletionPolicy(r xpv1.DeletionPolicy) {
	mg.Spec.DeletionPolicy = r
}

// SetProviderConfigReference of this MySQLServerVirtualNetworkRule.
func (mg *MySQLServerVirtualNetworkRule) SetProviderConfigReference(r *xpv1.Reference) {
	mg.Spec.ProviderConfigReference = r
}

/*
SetProviderReference of this MySQLServerVirtualNetworkRule.
Deprecated: Use SetProviderConfigReference.
*/
func (mg *MySQLServerVirtualNetworkRule) SetProviderReference(r *xpv1.Reference) {
	mg.Spec.ProviderReference = r
}

// SetPublishConnectionDetailsTo of this MySQLServerVirtualNetworkRule.
func (mg *MySQLServerVirtualNetworkRule) SetPublishConnectionDetailsTo(r *xpv1.PublishConnectionDetailsTo) {
	mg.Spec.PublishConnectionDetailsTo = r
}

// SetWriteConnectionSecretToReference of this MySQLServerVirtualNetworkRule.
func (mg *MySQLServerVirtualNetworkRule) SetWriteConnectionSecretToReference(r *xpv1.SecretReference) {
	mg.Spec.WriteConnectionSecretToReference = r
}

// GetCondition of this PostgreSQLServerFirewallRule.
func (mg *PostgreSQLServerFirewallRule) GetCondition(ct xpv1.ConditionType) xpv1.Condition {
	return mg.Status.GetCondition(ct)
}

// GetDeletionPolicy of this PostgreSQLServerFirewallRule.
func (mg *PostgreSQLServerFirewallRule) GetDeletionPolicy() xpv1.DeletionPolicy {
	return mg.Spec.DeletionPolicy
}

// GetProviderConfigReference of this PostgreSQLServerFirewallRule.
func (mg *PostgreSQLServerFirewallRule) GetProviderConfigReference() *xpv1.Reference {
	return mg.Spec.ProviderConfigReference
}

/*
GetProviderReference of this PostgreSQLServerFirewallRule.
Deprecated: Use GetProviderConfigReference.
*/
func (mg *PostgreSQLServerFirewallRule) GetProviderReference() *xpv1.Reference {
	return mg.Spec.ProviderReference
}

// GetPublishConnectionDetailsTo of this PostgreSQLServerFirewallRule.
func (mg *PostgreSQLServerFirewallRule) GetPublishConnectionDetailsTo() *xpv1.PublishConnectionDetailsTo {
	return mg.Spec.PublishConnectionDetailsTo
}

// GetWriteConnectionSecretToReference of this PostgreSQLServerFirewallRule.
func (mg *PostgreSQLServerFirewallRule) GetWriteConnectionSecretToReference() *xpv1.SecretReference {
	return mg.Spec.WriteConnectionSecretToReference
}

// SetConditions of this PostgreSQLServerFirewallRule.
func (mg *PostgreSQLServerFirewallRule) SetConditions(c ...xpv1.Condition) {
	mg.Status.SetConditions(c...)
}

// SetDeletionPolicy of this PostgreSQLServerFirewallRule.
func (mg *PostgreSQLServerFirewallRule) SetDeletionPolicy(r xpv1.DeletionPolicy) {
	mg.Spec.DeletionPolicy = r
}

// SetProviderConfigReference of this PostgreSQLServerFirewallRule.
func (mg *PostgreSQLServerFirewallRule) SetProviderConfigReference(r *xpv1.Reference) {
	mg.Spec.ProviderConfigReference = r
}

/*
SetProviderReference of this PostgreSQLServerFirewallRule.
Deprecated: Use SetProviderConfigReference.
*/
func (mg *PostgreSQLServerFirewallRule) SetProviderReference(r *xpv1.Reference) {
	mg.Spec.ProviderReference = r
}

// SetPublishConnectionDetailsTo of this PostgreSQLServerFirewallRule.
func (mg *PostgreSQLServerFirewallRule) SetPublishConnectionDetailsTo(r *xpv1.PublishConnectionDetailsTo) {
	mg.Spec.PublishConnectionDetailsTo = r
}

// SetWriteConnectionSecretToReference of this PostgreSQLServerFirewallRule.
func (mg *PostgreSQLServerFirewallRule) SetWriteConnectionSecretToReference(r *xpv1.SecretReference) {
	mg.Spec.WriteConnectionSecretToReference = r
}

// GetCondition of this PostgreSQLServerVirtualNetworkRule.
func (mg *PostgreSQLServerVirtualNetworkRule) GetCondition(ct xpv1.ConditionType) xpv1.Condition {
	return mg.Status.GetCondition(ct)
}

// GetDeletionPolicy of this PostgreSQLServerVirtualNetworkRule.
func (mg *PostgreSQLServerVirtualNetworkRule) GetDeletionPolicy() xpv1.DeletionPolicy {
	return mg.Spec.DeletionPolicy
}

// GetProviderConfigReference of this PostgreSQLServerVirtualNetworkRule.
func (mg *PostgreSQLServerVirtualNetworkRule) GetProviderConfigReference() *xpv1.Reference {
	return mg.Spec.ProviderConfigReference
}

/*
GetProviderReference of this PostgreSQLServerVirtualNetworkRule.
Deprecated: Use GetProviderConfigReference.
*/
func (mg *PostgreSQLServerVirtualNetworkRule) GetProviderReference() *xpv1.Reference {
	return mg.Spec.ProviderReference
}

// GetPublishConnectionDetailsTo of this PostgreSQLServerVirtualNetworkRule.
func (mg *PostgreSQLServerVirtualNetworkRule) GetPublishConnectionDetailsTo() *xpv1.PublishConnectionDetailsTo {
	return mg.Spec.PublishConnectionDetailsTo
}

// GetWriteConnectionSecretToReference of this PostgreSQLServerVirtualNetworkRule.
func (mg *PostgreSQLServerVirtualNetworkRule) GetWriteConnectionSecretToReference() *xpv1.SecretReference {
	return mg.Spec.WriteConnectionSecretToReference
}

// SetConditions of this PostgreSQLServerVirtualNetworkRule.
func (mg *PostgreSQLServerVirtualNetworkRule) SetConditions(c ...xpv1.Condition) {
	mg.Status.SetConditions(c...)
}

// SetDeletionPolicy of this PostgreSQLServerVirtualNetworkRule.
func (mg *PostgreSQLServerVirtualNetworkRule) SetDeletionPolicy(r xpv1.DeletionPolicy) {
	mg.Spec.DeletionPolicy = r
}

// SetProviderConfigReference of this PostgreSQLServerVirtualNetworkRule.
func (mg *PostgreSQLServerVirtualNetworkRule) SetProviderConfigReference(r *xpv1.Reference) {
	mg.Spec.ProviderConfigReference = r
}

/*
SetProviderReference of this PostgreSQLServerVirtualNetworkRule.
Deprecated: Use SetProviderConfigReference.
*/
func (mg *PostgreSQLServerVirtualNetworkRule) SetProviderReference(r *xpv1.Reference) {
	mg.Spec.ProviderReference = r
}

// SetPublishConnectionDetailsTo of this PostgreSQLServerVirtualNetworkRule.
func (mg *PostgreSQLServerVirtualNetworkRule) SetPublishConnectionDetailsTo(r *xpv1.PublishConnectionDetailsTo) {
	mg.Spec.PublishConnectionDetailsTo = r
}

// SetWriteConnectionSecretToReference of this PostgreSQLServerVirtualNetworkRule.
func (mg *PostgreSQLServerVirtualNetworkRule) SetWriteConnectionSecretToReference(r *xpv1.SecretReference) {
	mg.Spec.WriteConnectionSecretToReference = r
}