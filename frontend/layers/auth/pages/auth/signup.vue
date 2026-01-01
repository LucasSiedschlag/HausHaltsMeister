<template>
  <div class="mx-auto max-w-md space-y-6">
    <div>
      <h1 class="text-2xl font-semibold">Criar conta</h1>
      <p class="text-sm text-muted-foreground">Comece a organizar seu ledger.</p>
    </div>
    <form class="space-y-4">
      <div class="space-y-2">
        <label class="text-sm font-medium" for="display_name">Nome</label>
        <input id="display_name" type="text" class="w-full rounded-md border px-3 py-2" />
      </div>
      <div class="space-y-2">
        <label class="text-sm font-medium" for="email">Email</label>
        <input id="email" type="email" class="w-full rounded-md border px-3 py-2" />
      </div>
      <div class="space-y-2">
        <label class="text-sm font-medium" for="password">Senha</label>
        <input id="password" type="password" class="w-full rounded-md border px-3 py-2" />
      </div>
      <button type="submit" class="w-full rounded-md bg-primary px-3 py-2 text-sm font-medium text-primary-foreground">
        Criar conta
      </button>
    </form>
  </div>
</template>
