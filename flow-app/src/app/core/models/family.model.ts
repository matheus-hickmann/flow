export interface GroupMember {
  userId: string;
  displayName: string;
  role: 'owner' | 'member';
  joinedAt: string;
}

export interface Group {
  id: string;
  name: string;
  ownerId: string;
  createdAt: string;
  members?: GroupMember[];
}

export interface Invite {
  token: string;
  groupId: string;
  groupName: string;
  inviteeLabel: string;
  inviterName: string;
  expiresAt: string;
  status: 'pending' | 'accepted' | 'revoked';
}

export interface InvitePreview {
  groupName: string;
  inviterName: string;
  expiresAt: string;
  valid: boolean;
}

export interface SharedAccount {
  id: string;
  code: string;
  name: string;
  type: string;
  balance: number;
  color: string;
  shared: boolean;
  ownerId: string;
  ownerName: string;
}
