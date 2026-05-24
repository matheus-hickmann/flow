import { inject, Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, of, catchError, map } from 'rxjs';

import { ENVIRONMENT } from '../config';
import type { Group, Invite, InvitePreview, SharedAccount } from '../models/family.model';

@Injectable({ providedIn: 'root' })
export class FamilyService {
  private readonly http = inject(HttpClient);
  private readonly env = inject(ENVIRONMENT);

  private get base(): string {
    return `${this.env.apiUrl}/api/v1`;
  }

  // ── Groups ───────────────────────────────────────────────────────────────

  listGroups(): Observable<Group[]> {
    return this.http.get<Group[]>(`${this.base}/groups`).pipe(catchError(() => of([])));
  }

  getGroup(groupId: string): Observable<Group | null> {
    return this.http.get<Group>(`${this.base}/groups/${groupId}`).pipe(catchError(() => of(null)));
  }

  createGroup(name: string): Observable<Group> {
    return this.http.post<Group>(`${this.base}/groups`, { name });
  }

  deleteGroup(groupId: string): Observable<void> {
    return this.http.delete<void>(`${this.base}/groups/${groupId}`);
  }

  removeMember(groupId: string, userId: string): Observable<void> {
    return this.http.delete<void>(`${this.base}/groups/${groupId}/members/${userId}`);
  }

  listSharedAccounts(groupId: string): Observable<SharedAccount[]> {
    return this.http.get<SharedAccount[]>(`${this.base}/groups/${groupId}/accounts`).pipe(
      map((list) =>
        list.map((a) => ({ ...a, balance: parseBalance((a as { balance?: string | number }).balance) })),
      ),
      catchError(() => of([])),
    );
  }

  // ── Invites ───────────────────────────────────────────────────────────────

  listInvites(groupId: string): Observable<Invite[]> {
    return this.http.get<Invite[]>(`${this.base}/groups/${groupId}/invites`).pipe(catchError(() => of([])));
  }

  createInvite(groupId: string, inviteeLabel: string): Observable<Invite> {
    return this.http.post<Invite>(`${this.base}/groups/${groupId}/invites`, { inviteeLabel });
  }

  revokeInvite(groupId: string, token: string): Observable<void> {
    return this.http.delete<void>(`${this.base}/groups/${groupId}/invites/${token}`);
  }

  getInvitePreview(token: string): Observable<InvitePreview | null> {
    return this.http.get<InvitePreview>(`${this.base}/invites/${token}`).pipe(catchError(() => of(null)));
  }

  acceptInvite(token: string): Observable<void> {
    return this.http.post<void>(`${this.base}/invites/${token}/accept`, {});
  }
}

function parseBalance(val: number | string | undefined): number {
  if (val == null) return 0;
  if (typeof val === 'number') return val;
  const s = String(val).trim().replace(/\./g, '').replace(',', '.');
  return Number(s) || 0;
}
