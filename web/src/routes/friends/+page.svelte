<script lang="ts">
	import {
		listFriends,
		listIncoming,
		listOutgoing,
		sendRequest,
		acceptRequest,
		declineRequest,
		removeFriend,
		getCard,
		type Friend,
		type UserCard
	} from '$lib/api';
	import { user } from '$lib/user.svelte';
	import { toaster } from '$lib/toast.svelte';
	import { slide } from 'svelte/transition';

	const initials = (name: string) => name.slice(0, 2).toUpperCase();

	let friends = $state<Friend[]>([]);
	let incoming = $state<Friend[]>([]);
	let outgoing = $state<Friend[]>([]);
	let targetId = $state('');
	let match = $state<UserCard | null>(null);
	let lookupError = $state('');

	let lookupTimer: ReturnType<typeof setTimeout> | undefined;
	function onTargetInput() {
		match = null;
		lookupError = '';
		clearTimeout(lookupTimer);
		lookupTimer = setTimeout(lookup, 300);
	}

	async function lookup() {
		const id = Number(targetId.trim());
		if (!targetId.trim() || !Number.isInteger(id) || id <= 0) return;
		if (String(id) === user.id) {
			lookupError = "That's your own ID.";
			return;
		}
		try {
			match = await getCard(id);
		} catch {
			lookupError = 'No user with that ID.';
		}
	}

	$effect(() => {
		if (user.id) refresh();
	});

	async function refresh() {
		try {
			[friends, incoming, outgoing] = await Promise.all([
				listFriends(),
				listIncoming(),
				listOutgoing()
			]);
		} catch (e) {
			toaster.error(e instanceof Error ? e.message : String(e));
		}
	}

	async function act(fn: () => Promise<unknown>, ok: string) {
		try {
			await fn();
			toaster.success(ok);
			await refresh();
		} catch (e) {
			toaster.error(e instanceof Error ? e.message : String(e));
		}
	}

	function send() {
		if (!match) return;
		const m = match;
		act(() => sendRequest(m.id), `Request sent to ${m.username}.`).then(() => {
			targetId = '';
			match = null;
		});
	}
</script>

<h1>Friends</h1>
<p class="muted sub">Add friends to share your map with them.</p>

{#if !user.id}
	<div class="card empty">Loading…</div>
{:else}
	<section class="card">
		<h2>Add a friend</h2>
		<div class="row">
			<input
				class="input"
				type="text"
				inputmode="numeric"
				bind:value={targetId}
				oninput={onTargetInput}
				placeholder="friend's ID (a number)"
				onkeydown={(e) => e.key === 'Enter' && send()}
			/>
			<button class="btn" disabled={!match} onclick={send}>Send request</button>
		</div>
		{#if match}
			<p class="preview" transition:slide>
				<span class="avatar small-av">{initials(match.username)}</span>
				Found <strong>{match.username}</strong> — send them a request?
			</p>
		{:else if lookupError}
			<p class="preview err" transition:slide>{lookupError}</p>
		{/if}
	</section>

	{#if incoming.length > 0}
		<section>
			<h2>Incoming requests <span class="badge">{incoming.length}</span></h2>
			<ul>
				{#each incoming as f (f.id)}
					<li class="card" transition:slide>
						<span class="who">
							<span class="avatar">{initials(f.username)}</span>
							{f.username}
						</span>
						<span class="actions">
							<button class="btn" onclick={() => act(() => acceptRequest(f.id), 'Accepted.')}>
								Accept
							</button>
							<button
								class="btn btn-ghost"
								onclick={() => act(() => declineRequest(f.id), 'Declined.')}
							>
								Decline
							</button>
						</span>
					</li>
				{/each}
			</ul>
		</section>
	{/if}

	{#if outgoing.length > 0}
		<section>
			<h2>Sent requests <span class="badge">{outgoing.length}</span></h2>
			<ul>
				{#each outgoing as f (f.id)}
					<li class="card" transition:slide>
						<span class="who">
							<span class="avatar pending">{initials(f.username)}</span>
							{f.username}
							<span class="muted small">pending…</span>
						</span>
						<button
							class="btn btn-ghost"
							onclick={() => act(() => removeFriend(f.id), 'Cancelled.')}
						>
							Cancel
						</button>
					</li>
				{/each}
			</ul>
		</section>
	{/if}

	<section>
		<h2>Your friends <span class="badge">{friends.length}</span></h2>
		{#if friends.length === 0}
			<div class="card empty">No friends yet — send a request above.</div>
		{:else}
			<ul>
				{#each friends as f (f.id)}
					<li class="card" transition:slide>
						<a class="who" href={`/friends/${f.id}`}>
							<span class="avatar">{initials(f.username)}</span>
							{f.username}
						</a>
						<span class="actions">
							<a class="btn" href={`/friends/${f.id}`}>View map</a>
							<button
								class="btn btn-danger"
								onclick={() => act(() => removeFriend(f.id), 'Removed.')}
							>
								Unfriend
							</button>
						</span>
					</li>
				{/each}
			</ul>
		{/if}
	</section>
{/if}

<style>
	h1 {
		margin: 0;
	}
	.sub {
		margin: 0.25rem 0 1.25rem;
		font-size: 0.92rem;
	}
	section {
		margin: 1.25rem 0;
	}
	section.card {
		padding: 1.1rem 1.25rem;
	}
	h2 {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		margin: 0 0 0.75rem;
	}
	.row {
		display: flex;
		gap: 0.5rem;
	}
	.row .input {
		flex: 1;
		max-width: 380px;
	}
	.preview {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		margin: 0.75rem 0 0;
		font-size: 0.9rem;
		color: var(--text);
	}
	.preview.err {
		color: var(--danger);
	}
	.small-av {
		width: 26px;
		height: 26px;
		font-size: 0.7rem;
	}
	ul {
		list-style: none;
		padding: 0;
		margin: 0;
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
		max-width: 560px;
	}
	li {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 0.5rem;
		flex-wrap: wrap;
		padding: 0.6rem 0.85rem;
		transition:
			box-shadow 0.15s var(--ease),
			transform 0.15s var(--ease);
	}
	li:hover {
		box-shadow: var(--shadow);
		transform: translateY(-1px);
	}
	.who {
		display: flex;
		align-items: center;
		gap: 0.6rem;
		font-weight: 550;
		color: var(--text);
		text-decoration: none;
	}
	.avatar {
		display: grid;
		place-items: center;
		width: 34px;
		height: 34px;
		border-radius: 50%;
		background: linear-gradient(135deg, var(--primary), var(--primary-700));
		color: white;
		font-size: 0.78rem;
		font-weight: 600;
		flex-shrink: 0;
	}
	.avatar.pending {
		background: linear-gradient(135deg, #cbd5e1, #94a3b8);
	}
	.small {
		font-size: 0.78rem;
		font-weight: 400;
	}
	.actions {
		display: flex;
		gap: 0.4rem;
		align-items: center;
	}
	.empty {
		padding: 1.5rem;
		text-align: center;
		color: var(--muted);
	}
</style>
