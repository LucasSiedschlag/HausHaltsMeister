<script setup lang="ts">
import { Card, CardContent, CardHeader, CardTitle } from '@shared/components/ui/card'
import { Button } from '@shared/components/ui/button'
</script>

<template>
  <div class="grid gap-6">
    <section class="grid gap-3">
      <h1 class="text-2xl font-semibold">Saúde financeira do mês</h1>
      <p class="max-w-2xl text-sm text-muted-foreground">
        Acompanhe o saldo real, separando patrimônio próprio de valores transitórios.
      </p>
    </section>

    <section class="grid gap-4 md:grid-cols-3">
      <Card>
        <CardHeader>
          <CardTitle class="text-sm font-medium text-muted-foreground">Saldo disponível</CardTitle>
        </CardHeader>
        <CardContent>
          <p class="text-2xl font-semibold">R$ 12.480,00</p>
          <p class="text-xs text-muted-foreground">Após contas a pagar e receber</p>
        </CardContent>
      </Card>
      <Card>
        <CardHeader>
          <CardTitle class="text-sm font-medium text-muted-foreground">Orçamento do mês</CardTitle>
        </CardHeader>
        <CardContent>
          <p class="text-2xl font-semibold">R$ 4.200,00</p>
          <p class="text-xs text-muted-foreground">Planejado x realizado</p>
        </CardContent>
      </Card>
      <Card>
        <CardHeader>
          <CardTitle class="text-sm font-medium text-muted-foreground">Cartões ativos</CardTitle>
        </CardHeader>
        <CardContent>
          <p class="text-2xl font-semibold">3</p>
          <p class="text-xs text-muted-foreground">Próximo fechamento em 5 dias</p>
        </CardContent>
      </Card>
    </section>

    <section class="grid gap-4 md:grid-cols-2">
      <Card>
        <CardHeader>
          <CardTitle class="text-sm font-medium text-muted-foreground">Transações recentes</CardTitle>
        </CardHeader>
        <CardContent class="space-y-3 text-sm">
          <div class="flex items-center justify-between">
            <span>Mercado Central</span>
            <span class="text-destructive">- R$ 182,90</span>
          </div>
          <div class="flex items-center justify-between">
            <span>Salário</span>
            <span class="text-emerald-600">+ R$ 8.500,00</span>
          </div>
          <div class="flex items-center justify-between">
            <span>Assinaturas</span>
            <span class="text-destructive">- R$ 96,00</span>
          </div>
        </CardContent>
      </Card>
      <Card>
        <CardHeader>
          <CardTitle class="text-sm font-medium text-muted-foreground">Próximas ações</CardTitle>
        </CardHeader>
        <CardContent class="space-y-3 text-sm">
          <div class="flex items-center justify-between">
            <span>Fechamento Visa</span>
            <span class="text-muted-foreground">07/02</span>
          </div>
          <div class="flex items-center justify-between">
            <span>Revisar orçamento</span>
            <span class="text-muted-foreground">Em 3 dias</span>
          </div>
          <Button variant="outline" class="mt-2 w-full">Abrir painel</Button>
        </CardContent>
      </Card>
    </section>
  </div>
</template>
