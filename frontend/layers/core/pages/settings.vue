<script setup lang="ts">
import { AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle, AlertDialogTrigger } from '@shared/components/ui/alert-dialog'
import { Button } from '@shared/components/ui/button'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@shared/components/ui/card'
import { Input } from '@shared/components/ui/input'
import { Label } from '@shared/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@shared/components/ui/select'
import { Switch } from '@shared/components/ui/switch'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@shared/components/ui/tabs'
import { usePreferences } from '@shared/composables/usePreferences'
import { push } from 'notivue'
import { useAuth } from '#layers/auth/composables/useAuth'
import { useDebounceFn } from '@vueuse/core'

const { user, listSessions, revokeSession, logoutAll, logout, clearSession } = useAuth()
const router = useRouter()
const { preferences, fetchPreferences, updatePreferences } = usePreferences()

const savingSection = ref<'account' | 'notifications' | 'visual' | null>(null)
const autosaveReady = ref(false)
const autosaveAccountReady = ref(false)
const autosaveVisualReady = ref(false)
const isHydrating = ref(true)

const lastSavedAccount = reactive({
  locale: 'pt-BR',
  theme_mode: 'system'
})
const lastSavedNotifications = reactive({
  notify_card_close: true,
  notify_budget_over: true,
  notify_payables: true
})
const lastSavedVisual = reactive({
  compact_mode: 'comfortable',
  font_scale: 'md',
  theme_palette: 'default',
  theme_tone: 'vivid'
})

const activeTab = ref('conta')
const sessions = ref<Array<{
  id: string
  created_at: string
  expires_at: string
  user_agent?: string
  ip?: string
  device_name?: string
  is_current?: boolean
}>>([])
const sessionsLoading = ref(false)

const formatSessionDate = (value: string) => {
  if (!value) return ''
  return new Date(value).toLocaleString('pt-BR')
}

const sessionDeviceKey = (session: {
  user_agent?: string
  device_name?: string
  ip?: string
}) => {
  return session.device_name || session.user_agent || session.ip || 'unknown'
}

const latestSessionsByDevice = computed(() => {
  const seen = new Set<string>()
  const items: typeof sessions.value = []
  for (const session of sessions.value) {
    const key = sessionDeviceKey(session)
    if (seen.has(key)) continue
    seen.add(key)
    items.push(session)
  }
  return items
})

const form = reactive({
  displayName: '',
  email: '',
  locale: 'pt-BR',
  theme_mode: 'system',
  theme_palette: 'default',
  theme_tone: 'vivid',
  compact_mode: 'comfortable',
  font_scale: 'md',
  notify_card_close: true,
  notify_budget_over: true,
  notify_payables: true
})

const syncUser = () => {
  form.displayName = user.value?.display_name || ''
  form.email = user.value?.email || ''
}

const syncPreferences = () => {
  if (!preferences.value) return
  form.locale = preferences.value.locale
  form.theme_mode = preferences.value.theme_mode
  form.theme_palette = preferences.value.theme_palette
  form.theme_tone = preferences.value.theme_tone
  form.compact_mode = preferences.value.compact_mode
  form.font_scale = preferences.value.font_scale
  form.notify_card_close = preferences.value.notify_card_close
  form.notify_budget_over = preferences.value.notify_budget_over
  form.notify_payables = preferences.value.notify_payables

  lastSavedAccount.locale = preferences.value.locale
  lastSavedAccount.theme_mode = preferences.value.theme_mode
  lastSavedNotifications.notify_card_close = preferences.value.notify_card_close
  lastSavedNotifications.notify_budget_over = preferences.value.notify_budget_over
  lastSavedNotifications.notify_payables = preferences.value.notify_payables
  lastSavedVisual.compact_mode = preferences.value.compact_mode
  lastSavedVisual.font_scale = preferences.value.font_scale
  lastSavedVisual.theme_palette = preferences.value.theme_palette
  lastSavedVisual.theme_tone = preferences.value.theme_tone
}

const saveSection = async (
  section: 'account' | 'notifications' | 'visual',
  payload: {
    theme_mode?: string
    theme_palette?: string
    theme_tone?: string
    locale?: string
    compact_mode?: string
    font_scale?: string
    notify_card_close?: boolean
    notify_budget_over?: boolean
    notify_payables?: boolean
  }
) => {
  savingSection.value = section
  try {
    const updated = await updatePreferences(payload)
    if (updated) {
      push.success({
        title: 'Preferências',
        message: 'Preferências atualizadas.'
      })
      return updated
    }
    push.error({
      title: 'Preferências',
      message: 'Não foi possível salvar. Faça login novamente.'
    })
    return null
  } catch {
    push.error({
      title: 'Preferências',
      message: 'Não foi possível salvar. Tente novamente.'
    })
    return null
  } finally {
    savingSection.value = null
  }
}

const saveAccount = async () => {
  const updated = await saveSection('account', {
    locale: form.locale,
    theme_mode: form.theme_mode
  })
  if (updated) {
    lastSavedAccount.locale = form.locale
    lastSavedAccount.theme_mode = form.theme_mode
  }
}

const saveNotifications = async () => {
  const updated = await saveSection('notifications', {
    notify_card_close: form.notify_card_close,
    notify_budget_over: form.notify_budget_over,
    notify_payables: form.notify_payables
  })
  if (updated) {
    lastSavedNotifications.notify_card_close = form.notify_card_close
    lastSavedNotifications.notify_budget_over = form.notify_budget_over
    lastSavedNotifications.notify_payables = form.notify_payables
  }
}

const saveVisual = async () => {
  const updated = await saveSection('visual', {
    theme_palette: form.theme_palette,
    theme_tone: form.theme_tone,
    compact_mode: form.compact_mode,
    font_scale: form.font_scale
  })
  if (updated) {
    lastSavedVisual.theme_palette = form.theme_palette
    lastSavedVisual.theme_tone = form.theme_tone
    lastSavedVisual.compact_mode = form.compact_mode
    lastSavedVisual.font_scale = form.font_scale
  }
}

const isAccountDirty = () =>
  form.locale !== lastSavedAccount.locale || form.theme_mode !== lastSavedAccount.theme_mode

const isNotificationsDirty = () =>
  form.notify_card_close !== lastSavedNotifications.notify_card_close ||
  form.notify_budget_over !== lastSavedNotifications.notify_budget_over ||
  form.notify_payables !== lastSavedNotifications.notify_payables

const isVisualDirty = () =>
  form.compact_mode !== lastSavedVisual.compact_mode ||
  form.font_scale !== lastSavedVisual.font_scale ||
  form.theme_palette !== lastSavedVisual.theme_palette ||
  form.theme_tone !== lastSavedVisual.theme_tone

watch(user, syncUser, { immediate: true })
watch(preferences, syncPreferences, { immediate: true })

const autosaveAccount = useDebounceFn(() => {
  if (!autosaveAccountReady.value || isHydrating.value || !isAccountDirty()) return
  saveAccount()
}, 600)

const autosaveNotifications = useDebounceFn(() => {
  if (!autosaveReady.value || isHydrating.value || !isNotificationsDirty()) return
  saveNotifications()
}, 500)

const autosaveVisual = useDebounceFn(() => {
  if (!autosaveVisualReady.value || isHydrating.value || !isVisualDirty()) return
  saveVisual()
}, 600)

watch(
  () => [form.locale, form.theme_mode],
  () => {
    autosaveAccount()
  }
)

watch(
  () => [form.notify_card_close, form.notify_budget_over, form.notify_payables],
  () => {
    autosaveNotifications()
  }
)

watch(
  () => [form.compact_mode, form.font_scale, form.theme_palette, form.theme_tone],
  () => {
    autosaveVisual()
  }
)

onMounted(async () => {
  await fetchPreferences()
  autosaveReady.value = true
  autosaveAccountReady.value = true
  autosaveVisualReady.value = true
  await nextTick()
  isHydrating.value = false
})

const loadSessions = async () => {
  if (sessionsLoading.value) return
  sessionsLoading.value = true
  try {
    sessions.value = await listSessions()
  } catch {
    push.error({
      title: 'Sessões',
      message: 'Não foi possível carregar as sessões.'
    })
  } finally {
    sessionsLoading.value = false
  }
}

const handleRevokeSession = async (sessionId: string, isCurrent?: boolean) => {
  if (isCurrent) {
    try {
      await logout()
    } catch {
      clearSession()
    }
    await router.push('/auth/login')
    return
  }
  await revokeSession(sessionId)
  await loadSessions()
}

const handleLogoutAll = async () => {
  await logoutAll()
  clearSession()
  await router.push('/auth/login')
}

watch(activeTab, async (value) => {
  if (value === 'seguranca' && sessions.value.length === 0) {
    await loadSessions()
  }
})
</script>

<template>
  <div class="grid gap-6">
    <section class="grid gap-2">
      <h1 class="text-2xl font-semibold">Configurações</h1>
      <p class="text-sm text-muted-foreground">
        Preferências pessoais, segurança e ajustes do ledger.
      </p>
    </section>

    <Tabs v-model="activeTab" class="space-y-6">
      <TabsList class="flex w-full flex-wrap justify-start gap-2">
        <TabsTrigger value="conta">Conta</TabsTrigger>
        <TabsTrigger value="seguranca">Segurança</TabsTrigger>
        <TabsTrigger value="ledger">Ledger</TabsTrigger>
        <TabsTrigger value="notificacoes">Notificações</TabsTrigger>
        <TabsTrigger value="privacidade">Privacidade</TabsTrigger>
        <TabsTrigger value="visualizacao">Visualização</TabsTrigger>
      </TabsList>

      <TabsContent value="conta" class="space-y-6">
        <Card>
          <CardHeader>
            <CardTitle>Perfil</CardTitle>
            <CardDescription>Informações públicas e preferências de idioma.</CardDescription>
          </CardHeader>
          <CardContent class="grid gap-4 md:grid-cols-2">
            <div class="grid gap-2">
              <Label for="displayName">Nome exibido</Label>
              <Input id="displayName" v-model="form.displayName" placeholder="Seu nome" disabled />
            </div>
            <div class="grid gap-2">
              <Label for="email">E-mail</Label>
              <Input id="email" v-model="form.email" type="email" placeholder="email@exemplo.com" disabled />
            </div>
            <div class="grid gap-2">
              <Label>Idioma</Label>
              <Select v-model="form.locale">
                <SelectTrigger>
                  <SelectValue placeholder="Selecione" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="pt-BR">Português (Brasil)</SelectItem>
                  <SelectItem value="en-US">English (US)</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div class="grid gap-2">
              <Label>Tema</Label>
              <Select v-model="form.theme_mode">
                <SelectTrigger>
                  <SelectValue placeholder="Selecione" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="light">Claro</SelectItem>
                  <SelectItem value="dark">Escuro</SelectItem>
                  <SelectItem value="system">Sistema</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </CardContent>
        </Card>
      </TabsContent>

      <TabsContent value="seguranca" class="space-y-6">
        <Card>
          <CardHeader>
            <CardTitle>Segurança</CardTitle>
            <CardDescription>Senha, 2FA e sessões ativas.</CardDescription>
          </CardHeader>
          <CardContent class="grid gap-4 md:grid-cols-2">
            <div class="grid gap-2">
              <Label for="currentPassword">Senha atual</Label>
              <Input id="currentPassword" type="password" />
            </div>
            <div class="grid gap-2">
              <Label for="newPassword">Nova senha</Label>
              <Input id="newPassword" type="password" />
            </div>
            <div class="grid gap-2">
              <Label for="confirmPassword">Confirmar nova senha</Label>
              <Input id="confirmPassword" type="password" />
            </div>
            <div class="flex items-center justify-between rounded-lg border border-border/60 p-4">
              <div>
                <p class="text-sm font-medium">Autenticação em duas etapas</p>
                <p class="text-xs text-muted-foreground">Disponível em breve</p>
              </div>
              <Switch disabled />
            </div>
          </CardContent>
          <CardFooter class="flex justify-between">
            <AlertDialog>
              <AlertDialogTrigger as-child>
                <Button variant="outline">Encerrar sessões</Button>
              </AlertDialogTrigger>
              <AlertDialogContent>
                <AlertDialogHeader>
                  <AlertDialogTitle>Encerrar todas as sessões?</AlertDialogTitle>
                  <AlertDialogDescription>
                    Isso desconecta todos os dispositivos, incluindo este.
                  </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                  <AlertDialogCancel>Cancelar</AlertDialogCancel>
                  <AlertDialogAction @click="handleLogoutAll">Encerrar</AlertDialogAction>
                </AlertDialogFooter>
              </AlertDialogContent>
            </AlertDialog>
            <Button>Atualizar senha</Button>
          </CardFooter>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Sessões</CardTitle>
            <CardDescription>Dispositivos conectados recentemente.</CardDescription>
          </CardHeader>
          <CardContent class="space-y-3 text-sm">
            <div v-if="sessionsLoading" class="rounded-lg border border-border/60 p-3 text-muted-foreground">
              Carregando sessões...
            </div>
            <div v-else-if="latestSessionsByDevice.length === 0" class="rounded-lg border border-border/60 p-3 text-muted-foreground">
              Nenhuma sessão ativa.
            </div>
            <div
              v-for="session in latestSessionsByDevice"
              :key="session.id"
              class="flex items-center justify-between rounded-lg border border-border/60 p-3"
            >
              <div>
                <p class="font-medium">
                  {{ session.device_name || session.user_agent || 'Sessão' }}
                  <span v-if="session.is_current" class="ml-2 rounded-full bg-primary/10 px-2 py-0.5 text-xs text-primary">
                    Atual
                  </span>
                </p>
                <p class="text-xs text-muted-foreground">
                  Início: {{ formatSessionDate(session.created_at) }}
                </p>
              </div>
              <Button
                variant="ghost"
                size="sm"
                @click="handleRevokeSession(session.id, session.is_current)"
              >
                Sair
              </Button>
            </div>
          </CardContent>
        </Card>
      </TabsContent>

      <TabsContent value="ledger" class="space-y-6">
        <Card>
          <CardHeader>
            <CardTitle>Preferências do ledger</CardTitle>
            <CardDescription>Definições padrão para o seu ledger.</CardDescription>
          </CardHeader>
          <CardContent class="grid gap-4 md:grid-cols-2">
            <div class="grid gap-2">
              <Label for="currency">Moeda base</Label>
              <Select default-value="BRL">
                <SelectTrigger id="currency">
                  <SelectValue placeholder="Selecione" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="BRL">BRL</SelectItem>
                  <SelectItem value="USD">USD</SelectItem>
                  <SelectItem value="EUR">EUR</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div class="grid gap-2">
              <Label for="timezone">Fuso horário</Label>
              <Select default-value="America/Sao_Paulo">
                <SelectTrigger id="timezone">
                  <SelectValue placeholder="Selecione" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="America/Sao_Paulo">América/São Paulo</SelectItem>
                  <SelectItem value="UTC">UTC</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Contas</CardTitle>
            <CardDescription>Preferências para contas e cartões.</CardDescription>
          </CardHeader>
          <CardContent class="space-y-3">
            <div class="flex items-center justify-between rounded-lg border border-border/60 p-3">
              <div>
                <p class="font-medium">Conta principal padrão</p>
                <p class="text-xs text-muted-foreground">Selecionada ao criar transações.</p>
              </div>
              <Button variant="outline" size="sm">Selecionar</Button>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Membros e permissões</CardTitle>
            <CardDescription>Controle de acesso ao ledger.</CardDescription>
          </CardHeader>
          <CardContent class="space-y-3 text-sm">
            <div class="flex items-center justify-between rounded-lg border border-border/60 p-3">
              <div>
                <p class="font-medium">Lucas (owner)</p>
                <p class="text-xs text-muted-foreground">Você</p>
              </div>
              <Button variant="outline" size="sm">Gerenciar</Button>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Política de alterações</CardTitle>
            <CardDescription>Regra padrão para exclusões e edições.</CardDescription>
          </CardHeader>
          <CardContent>
            <div class="flex items-center justify-between rounded-lg border border-border/60 p-3">
              <div>
                <p class="font-medium">Ajustes obrigatórios</p>
                <p class="text-xs text-muted-foreground">Alterações via entradas de ajuste.</p>
              </div>
              <Button variant="outline" size="sm" disabled>Ativo</Button>
            </div>
          </CardContent>
        </Card>
      </TabsContent>

      <TabsContent value="notificacoes" class="space-y-6">
        <Card>
          <CardHeader>
            <CardTitle>Notificações</CardTitle>
            <CardDescription>Alertas essenciais do ledger.</CardDescription>
          </CardHeader>
          <CardContent class="space-y-3">
            <div class="flex items-center justify-between rounded-lg border border-border/60 p-3">
              <div>
                <p class="font-medium">Fechamento de cartão</p>
                <p class="text-xs text-muted-foreground">Avisar 3 dias antes do fechamento.</p>
              </div>
              <Switch v-model="form.notify_card_close" />
            </div>
            <div class="flex items-center justify-between rounded-lg border border-border/60 p-3">
              <div>
                <p class="font-medium">Orçamento estourado</p>
                <p class="text-xs text-muted-foreground">Alertas quando exceder o limite.</p>
              </div>
              <Switch v-model="form.notify_budget_over" />
            </div>
            <div class="flex items-center justify-between rounded-lg border border-border/60 p-3">
              <div>
                <p class="font-medium">Lembretes de contas a pagar</p>
                <p class="text-xs text-muted-foreground">Notificar 1 dia antes.</p>
              </div>
              <Switch v-model="form.notify_payables" />
            </div>
          </CardContent>
        </Card>
      </TabsContent>

      <TabsContent value="privacidade" class="space-y-6">
        <Card>
          <CardHeader>
            <CardTitle>Privacidade</CardTitle>
            <CardDescription>Controle de visibilidade de dados.</CardDescription>
          </CardHeader>
          <CardContent class="space-y-3">
            <div class="flex items-center justify-between rounded-lg border border-border/60 p-3">
              <div>
                <p class="font-medium">Exportar meus dados</p>
                <p class="text-xs text-muted-foreground">Disponível em breve.</p>
              </div>
              <Button variant="outline" size="sm" disabled>Exportar</Button>
            </div>
            <div class="flex items-center justify-between rounded-lg border border-border/60 p-3">
              <div>
                <p class="font-medium">Importar meus dados</p>
                <p class="text-xs text-muted-foreground">Disponível em breve.</p>
              </div>
              <Button variant="outline" size="sm" disabled>Importar</Button>
            </div>
          </CardContent>
        </Card>
      </TabsContent>

      <TabsContent value="visualizacao" class="space-y-6">
        <Card>
          <CardHeader>
            <CardTitle>Visualização</CardTitle>
            <CardDescription>Preferências de visualização.</CardDescription>
          </CardHeader>
          <CardContent class="grid gap-4 md:grid-cols-2">
            <div class="grid gap-2">
              <Label>Densidade</Label>
              <Select v-model="form.compact_mode">
                <SelectTrigger>
                  <SelectValue placeholder="Selecione" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="comfortable">Confortável</SelectItem>
                  <SelectItem value="compact">Compacta</SelectItem>
                  <SelectItem value="dense">Densa</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div class="grid gap-2">
              <Label>Tamanho da fonte</Label>
              <Select v-model="form.font_scale">
                <SelectTrigger>
                  <SelectValue placeholder="Selecione" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="sm">Pequena</SelectItem>
                  <SelectItem value="md">Média</SelectItem>
                  <SelectItem value="lg">Grande</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div class="grid gap-2">
              <Label>Paleta</Label>
              <Select v-model="form.theme_palette">
                <SelectTrigger>
                  <SelectValue placeholder="Selecione" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="default">Azul padrão</SelectItem>
                  <SelectItem value="red">Vermelho</SelectItem>
                  <SelectItem value="rose">Rose</SelectItem>
                  <SelectItem value="orange">Laranja</SelectItem>
                  <SelectItem value="green">Verde</SelectItem>
                  <SelectItem value="yellow">Amarelo</SelectItem>
                  <SelectItem value="violet">Violeta</SelectItem>
                  <SelectItem value="monochrome">Monocromático</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div class="grid gap-2">
              <Label>Tom</Label>
              <Select v-model="form.theme_tone">
                <SelectTrigger>
                  <SelectValue placeholder="Selecione" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="vivid">Vivo</SelectItem>
                  <SelectItem value="pastel">Pastel</SelectItem>
                  <SelectItem value="muted">Suave</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </CardContent>
        </Card>
      </TabsContent>
    </Tabs>
  </div>
</template>
