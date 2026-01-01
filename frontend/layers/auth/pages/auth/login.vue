<template>
  <div class="mx-auto max-w-md space-y-6">
    <div>
      <h1 class="text-2xl font-semibold">Login</h1>
      <p class="text-sm text-muted-foreground">Acesse seu ledger com suas credenciais.</p>
    </div>
    <form class="space-y-4">
      <div class="space-y-2">
        <label class="text-sm font-medium" for="email">Email</label>
        <input id="email" type="email" class="w-full rounded-md border px-3 py-2" />
      </div>
      <div class="space-y-2">
        <label class="text-sm font-medium" for="password">Senha</label>
        <input id="password" type="password" class="w-full rounded-md border px-3 py-2" />
      </div>
      <button type="submit" class="w-full rounded-md bg-primary px-3 py-2 text-sm font-medium text-primary-foreground">
        Entrar
      </button>
    </form>
    <div class="text-sm text-muted-foreground">
      <NuxtLink to="/auth/forgot-password" class="underline">Esqueci minha senha</NuxtLink>
    </div>
  </div>
</template>
