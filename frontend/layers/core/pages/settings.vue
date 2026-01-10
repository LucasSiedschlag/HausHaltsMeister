<script setup lang="ts">
import { AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle, AlertDialogTrigger } from '@shared/components/ui/alert-dialog'
import { Avatar, AvatarFallback, AvatarImage } from '@shared/components/ui/avatar'
import { Badge } from '@shared/components/ui/badge'
import { Button } from '@shared/components/ui/button'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@shared/components/ui/card'
import { Input } from '@shared/components/ui/input'
import { Label } from '@shared/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@shared/components/ui/select'
import { Switch } from '@shared/components/ui/switch'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@shared/components/ui/tabs'
import { useLedger } from '@shared/composables/useLedger'
import { usePreferences, type UserPreferences } from '@shared/composables/usePreferences'
import { useHeaderAction } from '@shared/composables/useHeaderAction'
import { push } from 'notivue'
import { useAuth } from '#layers/auth/composables/useAuth'
import { useLedgerMembers, type LedgerMember } from '#layers/ledgers/composables/useLedgerMembers'
import type { LedgerRole } from '#layers/shared/utils/ledger-roles'
import { useDebounceFn } from '@vueuse/core'
import { emailSchema, getInputClass, useInlineValidation } from '@shared/validators'

const { t, locale, setLocale } = useI18n()
const { setHeaderAction } = useHeaderAction()

const { user, listSessions, revokeSession, logoutAll, logout, clearSession } = useAuth()
const router = useRouter()
const { preferences, fetchPreferences, updatePreferences } = usePreferences()
const { ledgers, fetchLedgers, isLoading: ledgersLoading } = useLedger()
const { membersByLedger, loadingByLedger, fetchMembers, inviteMember, updateRole, removeMember } = useLedgerMembers()

setHeaderAction({ key: 'settings:save', labelKey: 'common.save', disabled: true })
type LocaleOption = 'pt-BR' | 'en-US'
type PreferencesUpdate = Partial<
  Pick<
    UserPreferences,
    | 'theme_mode'
    | 'theme_palette'
    | 'theme_tone'
    | 'locale'
    | 'compact_mode'
    | 'font_scale'
    | 'notify_card_close'
    | 'notify_budget_over'
    | 'notify_payables'
    | 'default_ledger_id'
  >
>

const savingSection = ref<'account' | 'notifications' | 'visual' | 'ledger' | null>(null)
const autosaveReady = ref(false)
const autosaveAccountReady = ref(false)
const autosaveVisualReady = ref(false)
const autosaveLedgerReady = ref(false)
const isHydrating = ref(true)
const isLocaleSwitching = ref(false)
const ledgerDataLoaded = ref(false)
const inviteDialogOpen = ref(false)
const inviteSaving = ref(false)

const lastSavedAccount = reactive<{ locale: LocaleOption; theme_mode: 'light' | 'dark' | 'system' }>({
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
const lastSavedLedger = reactive<{ default_ledger_id: string | null }>({
  default_ledger_id: null
})
const inviteForm = reactive<{ ledgerId: string; email: string; role: LedgerRole }>({
  ledgerId: '',
  email: '',
  role: 'viewer'
})
const inviteEmail = computed(() => inviteForm.email)
const inviteLedgerName = computed(() =>
  ledgers.value.find((ledger) => ledger.id === inviteForm.ledgerId)?.name ?? ''
)
const {
  touched: inviteTouched,
  errors: inviteErrors,
  touchField: touchInviteField,
  validateAll: validateInvite
} = useInlineValidation({ email: inviteEmail }, { email: emailSchema(t('settings.members.inviteEmail')) })

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
  return new Date(value).toLocaleString(locale.value)
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
  locale: 'pt-BR' as LocaleOption,
  theme_mode: 'system' as 'light' | 'dark' | 'system',
  theme_palette: 'default' as
    | 'default'
    | 'red'
    | 'rose'
    | 'orange'
    | 'green'
    | 'yellow'
    | 'violet'
    | 'monochrome',
  theme_tone: 'vivid' as 'vivid' | 'pastel' | 'muted',
  compact_mode: 'comfortable' as 'comfortable' | 'compact' | 'dense',
  font_scale: 'md' as 'sm' | 'md' | 'lg',
  notify_card_close: true,
  notify_budget_over: true,
  notify_payables: true,
  default_ledger_id: 'none' as string
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
  form.default_ledger_id = preferences.value.default_ledger_id ?? 'none'

  lastSavedAccount.locale = preferences.value.locale
  lastSavedAccount.theme_mode = preferences.value.theme_mode
  lastSavedNotifications.notify_card_close = preferences.value.notify_card_close
  lastSavedNotifications.notify_budget_over = preferences.value.notify_budget_over
  lastSavedNotifications.notify_payables = preferences.value.notify_payables
  lastSavedVisual.compact_mode = preferences.value.compact_mode
  lastSavedVisual.font_scale = preferences.value.font_scale
  lastSavedVisual.theme_palette = preferences.value.theme_palette
  lastSavedVisual.theme_tone = preferences.value.theme_tone
  lastSavedLedger.default_ledger_id = preferences.value.default_ledger_id
}

const saveSection = async (
  section: 'account' | 'notifications' | 'visual' | 'ledger',
  payload: PreferencesUpdate
) => {
  savingSection.value = section
  try {
    const updated = await updatePreferences(payload)
    if (updated) {
      push.success({
        title: t('settings.title'),
        message: t('settings.messages.updated')
      })
      return updated
    }
    push.error({
      title: t('settings.title'),
      message: t('settings.messages.sessionError')
    })
    return null
  } catch {
    push.error({
      title: t('settings.title'),
      message: t('settings.messages.saveError')
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

const resolveDefaultLedgerSelection = () =>
  form.default_ledger_id === 'none' ? null : form.default_ledger_id

const saveDefaultLedger = async () => {
  const updated = await saveSection('ledger', {
    default_ledger_id: resolveDefaultLedgerSelection()
  })
  if (updated) {
    lastSavedLedger.default_ledger_id = updated.default_ledger_id
  }
}

const isAccountDirty = () => form.theme_mode !== lastSavedAccount.theme_mode

const isNotificationsDirty = () =>
  form.notify_card_close !== lastSavedNotifications.notify_card_close ||
  form.notify_budget_over !== lastSavedNotifications.notify_budget_over ||
  form.notify_payables !== lastSavedNotifications.notify_payables

const isVisualDirty = () =>
  form.compact_mode !== lastSavedVisual.compact_mode ||
  form.font_scale !== lastSavedVisual.font_scale ||
  form.theme_palette !== lastSavedVisual.theme_palette ||
  form.theme_tone !== lastSavedVisual.theme_tone

const isLedgerDirty = () =>
  resolveDefaultLedgerSelection() !== lastSavedLedger.default_ledger_id

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

const autosaveLedger = useDebounceFn(() => {
  if (!autosaveLedgerReady.value || isHydrating.value || !isLedgerDirty()) return
  saveDefaultLedger()
}, 500)

watch(
  () => form.theme_mode,
  () => {
    autosaveAccount()
  }
)

watch(
  () => form.locale,
  async (value, previous) => {
    if (isHydrating.value || isLocaleSwitching.value) return
    if (value === previous) return
    if (value === lastSavedAccount.locale) return
    isLocaleSwitching.value = true
    const updated = await saveSection('account', { locale: value as LocaleOption })
    if (updated) {
      lastSavedAccount.locale = value
      await setLocale(value as LocaleOption)
    } else {
      form.locale = lastSavedAccount.locale
    }
    isLocaleSwitching.value = false
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

watch(
  () => form.default_ledger_id,
  () => {
    autosaveLedger()
  }
)

const loadLedgerData = async () => {
  if (ledgerDataLoaded.value) return
  await fetchLedgers()
  if (ledgers.value.length > 0) {
    await Promise.all(ledgers.value.map((ledger) => fetchMembers(ledger.id)))
  }
  ledgerDataLoaded.value = true
}

const membersForLedger = (ledgerId: string) => membersByLedger.value[ledgerId] ?? []
const membersLoading = (ledgerId: string) => Boolean(loadingByLedger.value[ledgerId])

const memberDisplayName = (member: LedgerMember) => member.display_name || member.email
const memberInitials = (member: LedgerMember) => {
  const source = memberDisplayName(member).trim()
  if (!source) return '??'
  const parts = source.split(' ').filter(Boolean)
  const first = parts[0]?.[0] ?? source[0]
  const second = parts.length > 1 ? parts[1]?.[0] : source[1]
  return `${first ?? ''}${second ?? ''}`.toUpperCase()
}

const openInviteDialog = (ledgerId: string) => {
  inviteForm.ledgerId = ledgerId
  inviteForm.email = ''
  inviteForm.role = 'viewer'
  inviteTouched.email = false
  inviteErrors.email = []
  inviteDialogOpen.value = true
}

const handleInviteMember = async () => {
  const email = inviteForm.email.trim()
  inviteForm.email = email
  if (!inviteForm.ledgerId) return
  if (validateInvite()) return
  inviteSaving.value = true
  try {
    await inviteMember(inviteForm.ledgerId, email.toLowerCase(), inviteForm.role)
    push.success({
      title: t('settings.members.inviteTitle'),
      message: t('settings.members.inviteSuccess')
    })
    inviteDialogOpen.value = false
  } catch {
    push.error({
      title: t('settings.members.inviteTitle'),
      message: t('settings.members.inviteError')
    })
  } finally {
    inviteSaving.value = false
  }
}

const handleRoleChange = async (ledgerId: string, member: LedgerMember, role: LedgerRole) => {
  if (role === member.role) return
  const previous = member.role
  member.role = role
  try {
    await updateRole(ledgerId, member.user_id, role)
    push.success({
      title: t('settings.members.updateTitle'),
      message: t('settings.members.updateSuccess')
    })
  } catch {
    member.role = previous
    push.error({
      title: t('settings.members.updateTitle'),
      message: t('settings.members.updateError')
    })
  }
}

const handleRemoveMember = async (ledgerId: string, member: LedgerMember) => {
  try {
    await removeMember(ledgerId, member.user_id)
    push.success({
      title: t('settings.members.removeTitle'),
      message: t('settings.members.removeSuccess')
    })
  } catch {
    push.error({
      title: t('settings.members.removeTitle'),
      message: t('settings.members.removeError')
    })
  }
}

onMounted(async () => {
  await fetchPreferences()
  autosaveReady.value = true
  autosaveAccountReady.value = true
  autosaveVisualReady.value = true
  autosaveLedgerReady.value = true
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
      title: t('settings.security.sessionsTitle'),
      message: t('settings.security.sessionsLoadError')
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
  if (value === 'ledger') {
    await loadLedgerData()
  }
})
</script>

<template>
  <div class="grid gap-6">
    <Tabs v-model="activeTab" class="space-y-6">
      <TabsList class="flex w-full flex-wrap justify-start gap-2">
        <TabsTrigger value="conta">{{ t('settings.tabs.account') }}</TabsTrigger>
        <TabsTrigger value="seguranca">{{ t('settings.tabs.security') }}</TabsTrigger>
        <TabsTrigger value="ledger">{{ t('settings.tabs.ledger') }}</TabsTrigger>
        <TabsTrigger value="notificacoes">{{ t('settings.tabs.notifications') }}</TabsTrigger>
        <TabsTrigger value="privacidade">{{ t('settings.tabs.privacy') }}</TabsTrigger>
        <TabsTrigger value="visualizacao">{{ t('settings.tabs.visual') }}</TabsTrigger>
      </TabsList>

      <TabsContent value="conta" class="space-y-6">
        <Card>
          <CardHeader>
            <CardTitle>{{ t('settings.profile.title') }}</CardTitle>
            <CardDescription>{{ t('settings.profile.description') }}</CardDescription>
          </CardHeader>
          <CardContent class="grid gap-4 md:grid-cols-2">
            <div class="grid gap-2">
              <Label for="displayName">{{ t('settings.profile.displayName') }}</Label>
              <Input id="displayName" v-model="form.displayName" :placeholder="t('settings.profile.displayNamePlaceholder')" disabled />
            </div>
            <div class="grid gap-2">
              <Label for="email">{{ t('settings.profile.email') }}</Label>
              <Input
                id="email"
                v-model="form.email"
                type="email"
                :placeholder="t('settings.profile.emailPlaceholder', { at: '@' })"
                disabled
              />
            </div>
            <div class="grid gap-2">
              <Label>{{ t('settings.locale') }}</Label>
              <Select v-model="form.locale">
                <SelectTrigger>
                  <SelectValue :placeholder="t('settings.common.selectPlaceholder')" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="pt-BR">{{ t('settings.locales.ptBR') }}</SelectItem>
                  <SelectItem value="en-US">{{ t('settings.locales.enUS') }}</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div class="grid gap-2">
              <Label>{{ t('settings.profile.theme') }}</Label>
              <Select v-model="form.theme_mode">
                <SelectTrigger>
                  <SelectValue :placeholder="t('settings.common.selectPlaceholder')" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="light">{{ t('settings.profile.themeOptions.light') }}</SelectItem>
                  <SelectItem value="dark">{{ t('settings.profile.themeOptions.dark') }}</SelectItem>
                  <SelectItem value="system">{{ t('settings.profile.themeOptions.system') }}</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </CardContent>
        </Card>
      </TabsContent>

      <TabsContent value="seguranca" class="space-y-6">
        <Card>
          <CardHeader>
            <CardTitle>{{ t('settings.sections.security') }}</CardTitle>
            <CardDescription>{{ t('settings.security.description') }}</CardDescription>
          </CardHeader>
          <CardContent class="grid gap-4 md:grid-cols-2">
            <div class="grid gap-2">
              <Label for="currentPassword">{{ t('settings.security.currentPassword') }}</Label>
              <Input id="currentPassword" type="password" />
            </div>
            <div class="grid gap-2">
              <Label for="newPassword">{{ t('settings.security.newPassword') }}</Label>
              <Input id="newPassword" type="password" />
            </div>
            <div class="grid gap-2">
              <Label for="confirmPassword">{{ t('settings.security.confirmPassword') }}</Label>
              <Input id="confirmPassword" type="password" />
            </div>
            <div class="flex items-center justify-between rounded-lg border border-border/60 p-4">
              <div>
                <p class="text-sm font-medium">{{ t('settings.security.twoFactor') }}</p>
                <p class="text-xs text-muted-foreground">{{ t('settings.common.comingSoon') }}</p>
              </div>
              <Switch disabled />
            </div>
          </CardContent>
          <CardFooter class="flex justify-between">
            <AlertDialog>
              <AlertDialogTrigger as-child>
                <Button variant="outline">{{ t('settings.security.endSessions') }}</Button>
              </AlertDialogTrigger>
              <AlertDialogContent>
                <AlertDialogHeader>
                  <AlertDialogTitle>{{ t('settings.security.endAllTitle') }}</AlertDialogTitle>
                  <AlertDialogDescription>
                    {{ t('settings.security.endAllDescription') }}
                  </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                  <AlertDialogCancel>{{ t('common.cancel') }}</AlertDialogCancel>
                  <AlertDialogAction @click="handleLogoutAll">{{ t('settings.security.endAllConfirm') }}</AlertDialogAction>
                </AlertDialogFooter>
              </AlertDialogContent>
            </AlertDialog>
            <Button>{{ t('settings.security.updatePassword') }}</Button>
          </CardFooter>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>{{ t('settings.security.sessionsTitle') }}</CardTitle>
            <CardDescription>{{ t('settings.security.sessionsDescription') }}</CardDescription>
          </CardHeader>
          <CardContent class="space-y-3 text-sm">
            <div v-if="sessionsLoading" class="rounded-lg border border-border/60 p-3 text-muted-foreground">
              {{ t('settings.security.sessionsLoading') }}
            </div>
            <div v-else-if="latestSessionsByDevice.length === 0" class="rounded-lg border border-border/60 p-3 text-muted-foreground">
              {{ t('settings.security.sessionsEmpty') }}
            </div>
            <div
              v-for="session in latestSessionsByDevice"
              :key="session.id"
              class="flex items-center justify-between rounded-lg border border-border/60 p-3"
            >
              <div>
                <p class="font-medium">
                  {{ session.device_name || session.user_agent || t('settings.security.sessionDefault') }}
                  <span v-if="session.is_current" class="ml-2 rounded-full bg-primary/10 px-2 py-0.5 text-xs text-primary">
                    {{ t('settings.security.sessionCurrent') }}
                  </span>
                </p>
                <p class="text-xs text-muted-foreground">
                  {{ t('settings.security.sessionStart') }}: {{ formatSessionDate(session.created_at) }}
                </p>
              </div>
              <Button
                variant="ghost"
                size="sm"
                @click="handleRevokeSession(session.id, session.is_current)"
              >
                {{ t('settings.security.sessionSignOut') }}
              </Button>
            </div>
          </CardContent>
        </Card>
      </TabsContent>

      <TabsContent value="ledger" class="space-y-6">
        <Card>
          <CardHeader>
            <CardTitle>{{ t('settings.ledger.title') }}</CardTitle>
            <CardDescription>{{ t('settings.ledger.description') }}</CardDescription>
          </CardHeader>
          <CardContent class="space-y-4">
            <div class="grid gap-2">
              <Label>{{ t('settings.ledger.defaultLedger') }}</Label>
              <Select v-model="form.default_ledger_id" :disabled="ledgersLoading">
                <SelectTrigger>
                  <SelectValue :placeholder="t('settings.ledger.defaultLedgerPlaceholder')" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="none">{{ t('settings.ledger.defaultLedgerNone') }}</SelectItem>
                  <SelectItem
                    v-for="ledger in ledgers"
                    :key="ledger.id"
                    :value="ledger.id"
                  >
                    {{ ledger.name }}
                  </SelectItem>
                </SelectContent>
              </Select>
              <p class="text-xs text-muted-foreground">
                {{ t('settings.ledger.defaultLedgerHint') }}
              </p>
            </div>
            <div v-if="ledgersLoading" class="rounded-lg border border-border/60 p-3 text-xs text-muted-foreground">
              {{ t('settings.ledger.loading') }}
            </div>
            <div
              v-else-if="ledgerDataLoaded && ledgers.length === 0"
              class="rounded-lg border border-border/60 p-3 text-xs text-muted-foreground"
            >
              {{ t('settings.ledger.empty') }}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>{{ t('settings.members.title') }}</CardTitle>
            <CardDescription>{{ t('settings.members.description') }}</CardDescription>
          </CardHeader>
          <CardContent class="space-y-4 text-sm">
            <div v-if="ledgersLoading && !ledgerDataLoaded" class="rounded-lg border border-border/60 p-3 text-muted-foreground">
              {{ t('settings.members.loadingLedgers') }}
            </div>
            <div v-else-if="ledgers.length === 0" class="rounded-lg border border-border/60 p-3 text-muted-foreground">
              {{ t('settings.members.noLedgers') }}
            </div>
            <div v-else class="space-y-4">
              <div
                v-for="ledger in ledgers"
                :key="ledger.id"
                class="space-y-3 rounded-lg border border-border/60 p-4"
              >
                <div class="flex flex-wrap items-start justify-between gap-3">
                  <div>
                    <p class="text-sm font-medium">{{ ledger.name }}</p>
                    <p class="text-xs text-muted-foreground">
                      {{ t('settings.members.ledgerRole', { role: t(`ledgers.roles.${ledger.role ?? 'unknown'}`) }) }}
                    </p>
                  </div>
                  <Badge variant="outline" class="text-xs">
                    {{ t(`ledgers.roles.${ledger.role ?? 'unknown'}`) }}
                  </Badge>
                </div>

                <div v-if="membersLoading(ledger.id)" class="rounded-lg border border-border/60 p-3 text-xs text-muted-foreground">
                  {{ t('settings.members.loading') }}
                </div>
                <div
                  v-else-if="membersForLedger(ledger.id).length === 0"
                  class="rounded-lg border border-border/60 p-3 text-xs text-muted-foreground"
                >
                  {{ t('settings.members.empty') }}
                </div>
                <div v-else class="space-y-2">
                  <div
                    v-for="member in membersForLedger(ledger.id)"
                    :key="member.user_id"
                    class="flex flex-wrap items-center justify-between gap-3 rounded-lg border border-border/60 p-3"
                  >
                    <div class="flex items-center gap-3">
                      <Avatar class="h-9 w-9">
                        <AvatarImage :src="member.avatar_url || ''" :alt="memberDisplayName(member)" />
                        <AvatarFallback class="bg-primary/10 text-xs text-primary">
                          {{ memberInitials(member) }}
                        </AvatarFallback>
                      </Avatar>
                      <div>
                        <p class="text-sm font-medium">
                          {{ memberDisplayName(member) }}
                          <span v-if="member.user_id === user?.id" class="ml-2 text-xs text-muted-foreground">
                            {{ t('settings.members.you') }}
                          </span>
                        </p>
                        <p class="text-xs text-muted-foreground">{{ member.email }}</p>
                      </div>
                    </div>
                    <div class="flex items-center gap-2">
                      <Badge v-if="member.role === 'owner'" variant="outline" class="text-xs">
                        {{ t('ledgers.roles.owner') }}
                      </Badge>
                      <Select
                        v-else
                        :model-value="member.role"
                        :disabled="ledger.role !== 'owner'"
                        @update:model-value="(value) => handleRoleChange(ledger.id, member, value as LedgerRole)"
                      >
                        <SelectTrigger class="h-8 w-[120px]">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="editor">{{ t('ledgers.roles.editor') }}</SelectItem>
                          <SelectItem value="viewer">{{ t('ledgers.roles.viewer') }}</SelectItem>
                        </SelectContent>
                      </Select>
                      <Button
                        variant="ghost"
                        size="sm"
                        :disabled="ledger.role !== 'owner' || member.role === 'owner'"
                        @click="handleRemoveMember(ledger.id, member)"
                      >
                        {{ t('settings.members.remove') }}
                      </Button>
                    </div>
                  </div>
                </div>

                <div class="flex justify-end">
                  <Button
                    variant="outline"
                    size="sm"
                    :disabled="ledger.role !== 'owner'"
                    @click="openInviteDialog(ledger.id)"
                  >
                    {{ t('settings.members.invite') }}
                  </Button>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>{{ t('settings.sections.accounts') }}</CardTitle>
            <CardDescription>{{ t('settings.accounts.description') }}</CardDescription>
          </CardHeader>
          <CardContent class="space-y-3">
            <div class="flex items-center justify-between rounded-lg border border-border/60 p-3">
              <div>
                <p class="font-medium">{{ t('settings.ledger.defaultAccount') }}</p>
                <p class="text-xs text-muted-foreground">{{ t('settings.accounts.defaultAccountHint') }}</p>
              </div>
              <Button variant="outline" size="sm">{{ t('settings.accounts.select') }}</Button>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>{{ t('settings.changePolicy.title') }}</CardTitle>
            <CardDescription>{{ t('settings.changePolicy.description') }}</CardDescription>
          </CardHeader>
          <CardContent>
            <div class="flex items-center justify-between rounded-lg border border-border/60 p-3">
              <div>
                <p class="font-medium">{{ t('settings.changePolicy.requiredAdjustments') }}</p>
                <p class="text-xs text-muted-foreground">{{ t('settings.changePolicy.requiredAdjustmentsHint') }}</p>
              </div>
              <Button variant="outline" size="sm" disabled>{{ t('settings.common.active') }}</Button>
            </div>
          </CardContent>
        </Card>

        <AlertDialog v-model:open="inviteDialogOpen">
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>{{ t('settings.members.inviteTitle') }}</AlertDialogTitle>
              <AlertDialogDescription>
                {{ t('settings.members.inviteDescription', { ledger: inviteLedgerName }) }}
              </AlertDialogDescription>
            </AlertDialogHeader>
            <div class="grid gap-4 py-4">
              <div class="grid gap-2">
                <Label for="inviteEmail">{{ t('settings.members.inviteEmail') }}</Label>
                <Input
                  id="inviteEmail"
                  v-model="inviteForm.email"
                  :class="getInputClass({ touched: inviteTouched.email, hasError: Boolean(inviteErrors.email[0]) })"
                  type="email"
                  :placeholder="t('settings.members.inviteEmailPlaceholder', { at: '@' })"
                  @blur="touchInviteField('email')"
                />
                <p v-if="inviteTouched.email && inviteErrors.email[0]" class="text-xs text-destructive">
                  {{ inviteErrors.email[0].message }}
                </p>
              </div>
              <div class="grid gap-2">
                <Label>{{ t('settings.members.inviteRole') }}</Label>
                <Select v-model="inviteForm.role">
                  <SelectTrigger>
                    <SelectValue :placeholder="t('settings.common.selectPlaceholder')" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="editor">{{ t('ledgers.roles.editor') }}</SelectItem>
                    <SelectItem value="viewer">{{ t('ledgers.roles.viewer') }}</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
            <AlertDialogFooter>
              <AlertDialogCancel>{{ t('common.cancel') }}</AlertDialogCancel>
              <Button
                :disabled="inviteSaving || !inviteForm.email.trim() || inviteErrors.email.length > 0"
                @click="handleInviteMember"
              >
                {{ inviteSaving ? t('settings.members.inviteSending') : t('settings.members.inviteConfirm') }}
              </Button>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      </TabsContent>

      <TabsContent value="notificacoes" class="space-y-6">
        <Card>
          <CardHeader>
            <CardTitle>{{ t('settings.sections.notifications') }}</CardTitle>
            <CardDescription>{{ t('settings.notifications.description') }}</CardDescription>
          </CardHeader>
          <CardContent class="space-y-3">
            <div class="flex items-center justify-between rounded-lg border border-border/60 p-3">
              <div>
                <p class="font-medium">{{ t('settings.notifications.cardCloseTitle') }}</p>
                <p class="text-xs text-muted-foreground">{{ t('settings.notifications.cardCloseDesc') }}</p>
              </div>
              <Switch v-model="form.notify_card_close" />
            </div>
            <div class="flex items-center justify-between rounded-lg border border-border/60 p-3">
              <div>
                <p class="font-medium">{{ t('settings.notifications.budgetOverTitle') }}</p>
                <p class="text-xs text-muted-foreground">{{ t('settings.notifications.budgetOverDesc') }}</p>
              </div>
              <Switch v-model="form.notify_budget_over" />
            </div>
            <div class="flex items-center justify-between rounded-lg border border-border/60 p-3">
              <div>
                <p class="font-medium">{{ t('settings.notifications.payablesTitle') }}</p>
                <p class="text-xs text-muted-foreground">{{ t('settings.notifications.payablesDesc') }}</p>
              </div>
              <Switch v-model="form.notify_payables" />
            </div>
          </CardContent>
        </Card>
      </TabsContent>

      <TabsContent value="privacidade" class="space-y-6">
        <Card>
          <CardHeader>
            <CardTitle>{{ t('settings.sections.privacy') }}</CardTitle>
            <CardDescription>{{ t('settings.privacy.description') }}</CardDescription>
          </CardHeader>
          <CardContent class="space-y-3">
            <div class="flex items-center justify-between rounded-lg border border-border/60 p-3">
              <div>
                <p class="font-medium">{{ t('settings.privacy.exportTitle') }}</p>
                <p class="text-xs text-muted-foreground">{{ t('settings.common.comingSoon') }}</p>
              </div>
              <Button variant="outline" size="sm" disabled>{{ t('settings.privacy.exportAction') }}</Button>
            </div>
            <div class="flex items-center justify-between rounded-lg border border-border/60 p-3">
              <div>
                <p class="font-medium">{{ t('settings.privacy.importTitle') }}</p>
                <p class="text-xs text-muted-foreground">{{ t('settings.common.comingSoon') }}</p>
              </div>
              <Button variant="outline" size="sm" disabled>{{ t('settings.privacy.importAction') }}</Button>
            </div>
          </CardContent>
        </Card>
      </TabsContent>

      <TabsContent value="visualizacao" class="space-y-6">
        <Card>
          <CardHeader>
            <CardTitle>{{ t('settings.sections.visual') }}</CardTitle>
            <CardDescription>{{ t('settings.visual.description') }}</CardDescription>
          </CardHeader>
          <CardContent class="grid gap-4 md:grid-cols-2">
            <div class="grid gap-2">
              <Label>{{ t('settings.visual.density') }}</Label>
              <Select v-model="form.compact_mode">
                <SelectTrigger>
                  <SelectValue :placeholder="t('settings.common.selectPlaceholder')" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="comfortable">{{ t('settings.visual.densityOptions.comfortable') }}</SelectItem>
                  <SelectItem value="compact">{{ t('settings.visual.densityOptions.compact') }}</SelectItem>
                  <SelectItem value="dense">{{ t('settings.visual.densityOptions.dense') }}</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div class="grid gap-2">
              <Label>{{ t('settings.visual.fontScale') }}</Label>
              <Select v-model="form.font_scale">
                <SelectTrigger>
                  <SelectValue :placeholder="t('settings.common.selectPlaceholder')" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="sm">{{ t('settings.visual.fontOptions.small') }}</SelectItem>
                  <SelectItem value="md">{{ t('settings.visual.fontOptions.medium') }}</SelectItem>
                  <SelectItem value="lg">{{ t('settings.visual.fontOptions.large') }}</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div class="grid gap-2">
              <Label>{{ t('settings.visual.palette') }}</Label>
              <Select v-model="form.theme_palette">
                <SelectTrigger>
                  <SelectValue :placeholder="t('settings.common.selectPlaceholder')" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="default">{{ t('settings.visual.paletteOptions.default') }}</SelectItem>
                  <SelectItem value="red">{{ t('settings.visual.paletteOptions.red') }}</SelectItem>
                  <SelectItem value="rose">{{ t('settings.visual.paletteOptions.rose') }}</SelectItem>
                  <SelectItem value="orange">{{ t('settings.visual.paletteOptions.orange') }}</SelectItem>
                  <SelectItem value="green">{{ t('settings.visual.paletteOptions.green') }}</SelectItem>
                  <SelectItem value="yellow">{{ t('settings.visual.paletteOptions.yellow') }}</SelectItem>
                  <SelectItem value="violet">{{ t('settings.visual.paletteOptions.violet') }}</SelectItem>
                  <SelectItem value="monochrome">{{ t('settings.visual.paletteOptions.monochrome') }}</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div class="grid gap-2">
              <Label>{{ t('settings.visual.tone') }}</Label>
              <Select v-model="form.theme_tone">
                <SelectTrigger>
                  <SelectValue :placeholder="t('settings.common.selectPlaceholder')" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="vivid">{{ t('settings.visual.toneOptions.vivid') }}</SelectItem>
                  <SelectItem value="pastel">{{ t('settings.visual.toneOptions.pastel') }}</SelectItem>
                  <SelectItem value="muted">{{ t('settings.visual.toneOptions.muted') }}</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </CardContent>
        </Card>
      </TabsContent>
    </Tabs>
  </div>
</template>
