<template>
  <div class="space-y-6">
    <div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
      <div class="space-y-1">
        <h1 class="text-2xl font-semibold tracking-tight">Dashboard</h1>
        <p class="text-sm text-muted-foreground">
          Overview of your household finance activity.
        </p>
      </div>
      <div class="flex flex-wrap items-center gap-2">
        <Button variant="outline">Export</Button>
        <Button>Add entry</Button>
      </div>
    </div>

    <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
      <div class="rounded-lg border bg-card p-4 text-card-foreground shadow-sm">
        <p class="text-sm text-muted-foreground">Monthly balance</p>
        <p class="mt-2 text-2xl font-semibold">$4,320</p>
        <p class="mt-2 text-xs text-muted-foreground">+12.4% vs last month</p>
      </div>
      <div class="rounded-lg border bg-card p-4 text-card-foreground shadow-sm">
        <p class="text-sm text-muted-foreground">Upcoming bills</p>
        <p class="mt-2 text-2xl font-semibold">$1,280</p>
        <p class="mt-2 text-xs text-muted-foreground">3 payments scheduled</p>
      </div>
      <div class="rounded-lg border bg-card p-4 text-card-foreground shadow-sm">
        <p class="text-sm text-muted-foreground">Savings goal</p>
        <p class="mt-2 text-2xl font-semibold">62%</p>
        <p class="mt-2 text-xs text-muted-foreground">Target by Dec 2024</p>
      </div>
    </div>

    <div class="rounded-lg border bg-card p-6 text-card-foreground shadow-sm">
      <div class="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
        <div>
          <h2 class="text-lg font-semibold">Recent activity</h2>
          <p class="text-sm text-muted-foreground">
            Last 30 days of transactions and updates.
          </p>
        </div>
        <Button variant="secondary">View reports</Button>
      </div>
      <div class="mt-4 grid gap-3 md:grid-cols-2">
        <div class="rounded-md border border-dashed bg-muted/30 p-4">
          <p class="text-sm font-medium">Utilities</p>
          <p class="mt-1 text-sm text-muted-foreground">$320 paid on Aug 12</p>
        </div>
        <div class="rounded-md border border-dashed bg-muted/30 p-4">
          <p class="text-sm font-medium">Groceries</p>
          <p class="mt-1 text-sm text-muted-foreground">$210 paid on Aug 10</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
</script>
