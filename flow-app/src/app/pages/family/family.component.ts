import { Component, inject, OnInit, signal, computed } from '@angular/core';
import { NgClass, SlicePipe } from '@angular/common';

import { FamilyService } from '../../core/services/family.service';
import { AuthService } from '../../core/services/auth.service';
import type { Group, Invite, SharedAccount } from '../../core/models/family.model';
import { CurrencyBrlPipe } from '../../shared';

@Component({
  selector: 'app-family',
  standalone: true,
  imports: [NgClass, SlicePipe, CurrencyBrlPipe],
  templateUrl: './family.component.html',
})
export class FamilyComponent implements OnInit {
  private readonly familyService = inject(FamilyService);
  private readonly authService = inject(AuthService);

  readonly loading = signal(true);
  readonly error = signal<string | null>(null);

  readonly groups = signal<Group[]>([]);
  readonly selectedGroup = signal<Group | null>(null);
  readonly invites = signal<Invite[]>([]);
  readonly sharedAccounts = signal<SharedAccount[]>([]);

  // Create group
  readonly showCreateGroup = signal(false);
  readonly newGroupName = signal('');
  readonly createGroupError = signal<string | null>(null);

  // Create invite
  readonly showCreateInvite = signal(false);
  readonly newInviteeLabel = signal('');
  readonly createInviteError = signal<string | null>(null);
  readonly createdInviteLink = signal<string | null>(null);

  // Derived
  readonly currentUserId = computed(() => this.authService.user()?.userId ?? '');
  readonly pendingInvites = computed(() => this.invites().filter((i) => i.status === 'pending'));

  ngOnInit(): void {
    this.loadGroups();
  }

  loadGroups(): void {
    this.loading.set(true);
    this.familyService.listGroups().subscribe({
      next: (groups) => {
        this.groups.set(groups);
        this.loading.set(false);
        if (groups.length > 0 && !this.selectedGroup()) {
          this.selectGroup(groups[0]);
        }
      },
      error: () => {
        this.error.set('Erro ao carregar grupos.');
        this.loading.set(false);
      },
    });
  }

  selectGroup(group: Group): void {
    this.familyService.getGroup(group.id).subscribe({
      next: (full) => {
        this.selectedGroup.set(full);
        this.loadInvites(group.id);
        this.loadSharedAccounts(group.id);
      },
    });
  }

  private loadInvites(groupId: string): void {
    this.familyService.listInvites(groupId).subscribe({
      next: (invites) => this.invites.set(invites),
      error: () => this.invites.set([]),
    });
  }

  private loadSharedAccounts(groupId: string): void {
    this.familyService.listSharedAccounts(groupId).subscribe({
      next: (accounts) => this.sharedAccounts.set(accounts),
      error: () => this.sharedAccounts.set([]),
    });
  }

  // ── Create group ─────────────────────────────────────────────────────────

  openCreateGroup(): void {
    this.newGroupName.set('');
    this.createGroupError.set(null);
    this.showCreateGroup.set(true);
  }

  submitCreateGroup(): void {
    const name = this.newGroupName().trim();
    if (!name) {
      this.createGroupError.set('Informe um nome para o grupo.');
      return;
    }
    this.familyService.createGroup(name).subscribe({
      next: (group) => {
        this.showCreateGroup.set(false);
        this.groups.update((gs) => [...gs, group]);
        this.selectGroup(group);
      },
      error: () => this.createGroupError.set('Erro ao criar grupo.'),
    });
  }

  deleteGroup(group: Group): void {
    if (!confirm(`Excluir o grupo "${group.name}"? Esta ação não pode ser desfeita.`)) return;
    this.familyService.deleteGroup(group.id).subscribe({
      next: () => {
        this.groups.update((gs) => gs.filter((g) => g.id !== group.id));
        if (this.selectedGroup()?.id === group.id) {
          const remaining = this.groups();
          this.selectedGroup.set(remaining[0] ?? null);
          if (remaining[0]) this.selectGroup(remaining[0]);
          else {
            this.invites.set([]);
            this.sharedAccounts.set([]);
          }
        }
      },
    });
  }

  removeMember(userId: string): void {
    const group = this.selectedGroup();
    if (!group) return;
    if (!confirm('Remover este membro do grupo?')) return;
    this.familyService.removeMember(group.id, userId).subscribe({
      next: () => this.selectGroup(group),
    });
  }

  // ── Invites ───────────────────────────────────────────────────────────────

  openCreateInvite(): void {
    this.newInviteeLabel.set('');
    this.createInviteError.set(null);
    this.createdInviteLink.set(null);
    this.showCreateInvite.set(true);
  }

  submitCreateInvite(): void {
    const group = this.selectedGroup();
    if (!group) return;
    const label = this.newInviteeLabel().trim();
    this.familyService.createInvite(group.id, label).subscribe({
      next: (invite) => {
        const link = `${window.location.origin}/entrar/${invite.token}`;
        this.createdInviteLink.set(link);
        this.invites.update((list) => [...list, invite]);
      },
      error: () => this.createInviteError.set('Erro ao gerar convite.'),
    });
  }

  copyLink(link: string): void {
    navigator.clipboard.writeText(link).catch(() => {});
  }

  revokeInvite(invite: Invite): void {
    const group = this.selectedGroup();
    if (!group) return;
    this.familyService.revokeInvite(group.id, invite.token).subscribe({
      next: () => this.loadInvites(group.id),
    });
  }

  isOwner(): boolean {
    return this.selectedGroup()?.ownerId === this.currentUserId();
  }

  trackByGroupId(_: number, g: Group): string { return g.id; }
  trackByUserId(_: number, m: { userId: string }): string { return m.userId; }
  trackByToken(_: number, i: Invite): string { return i.token; }
  trackByAccountId(_: number, a: SharedAccount): string { return a.id; }
}
