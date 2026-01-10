<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { Crown, Eye, Pencil } from 'lucide-vue-next'
import { Button } from '@shared/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@shared/components/ui/card'
import { Input } from '@shared/components/ui/input'
import { Label } from '@shared/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@shared/components/ui/select'
import { useLedger } from '@shared/composables/useLedger'
import { useHeaderAction } from '@shared/composables/useHeaderAction'
import { useAuth } from '#layers/auth/composables/useAuth'
import type { LedgerRole } from '#layers/shared/utils/ledger-roles'
import { push } from 'notivue'

const { t } = useI18n()
const router = useRouter()
const { ledgers, currentLedgerId, fetchLedgers, selectLedger, createLedger, isLoading } = useLedger()
const { user } = useAuth()
const { setHeaderAction } = useHeaderAction()

const form = reactive({
  name: '',
  currency_code: 'BRL'
})
const isCreating = ref(false)
const createCard = ref<{ $el: HTMLElement } | null>(null)
const nameInput = ref<{ $el: HTMLInputElement } | null>(null)

const defaultLedgerName = computed(() => {
  const source = user.value?.display_name || user.value?.email || ''
  const fallback = source.split('@')[0] || ''
  const first = fallback.trim().split(/\s+/)[0] || ''
  if (!first) return ''
  return `Conta do ${first}`
})

const ownedLedgers = computed(() => ledgers.value.filter((ledger) => ledger.role === 'owner'))
const sharedLedgers = computed(() => ledgers.value.filter((ledger) => ledger.role !== 'owner'))

const roleMeta = (role?: LedgerRole | null) => {
  if (role === 'owner') {
    return { label: t('ledgers.roles.owner'), icon: Crown, class: 'border-emerald-500/30 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300' }
  }
  if (role === 'editor') {
    return { label: t('ledgers.roles.editor'), icon: Pencil, class: 'border-sky-500/30 bg-sky-500/10 text-sky-700 dark:text-sky-300' }
  }
  if (role === 'viewer') {
    return { label: t('ledgers.roles.viewer'), icon: Eye, class: 'border-muted-foreground/30 bg-muted/50 text-muted-foreground' }
  }
  return { label: t('ledgers.roles.unknown'), icon: Eye, class: 'border-muted-foreground/30 bg-muted/50 text-muted-foreground' }
}

const loadLedgers = async () => {
  await fetchLedgers()
}

const handleSelect = async (ledgerId: string) => {
  await selectLedger(ledgerId)
  await router.push('/')
}

const handleCreate = async () => {
  if (!form.name.trim()) {
    push.error({
      title: t('ledgers.create.title'),
      message: t('ledgers.create.nameRequired')
    })
    return
  }
  isCreating.value = true
  try {
    const created = await createLedger({ name: form.name.trim(), currency_code: form.currency_code })
    if (created) {
      await router.push('/')
      return
    }
    push.error({
      title: t('ledgers.create.title'),
      message: t('ledgers.create.error')
    })
  } catch {
    push.error({
      title: t('ledgers.create.title'),
      message: t('ledgers.create.error')
    })
  } finally {
    isCreating.value = false
  }
}

const focusCreateLedger = () => {
  const target = createCard.value?.$el
  if (target) {
    target.scrollIntoView({ behavior: 'smooth', block: 'start' })
  }
  nameInput.value?.$el?.focus()
}

setHeaderAction(
  { key: 'ledgers:new', labelKey: 'ledgers.actions.new' },
  focusCreateLedger
)


onMounted(async () => {
  await loadLedgers()
})

watch(
  () => defaultLedgerName.value,
  (value) => {
    if (!form.name.trim() && value) {
      form.name = value
    }
  },
  { immediate: true }
)
</script>

<template>
  <div class="space-y-6">
    <div class="grid gap-4 lg:grid-cols-[1.4fr_1fr]">
      <Card ref="createCard">
        <CardHeader>
          <CardTitle>{{ t('ledgers.list.title') }}</CardTitle>
          <CardDescription>{{ t('ledgers.list.description') }}</CardDescription>
        </CardHeader>
        <CardContent class="space-y-4">
          <div v-if="isLoading" class="text-sm text-muted-foreground">
            {{ t('ledgers.loading') }}
          </div>
          <div v-else-if="ledgers.length === 0" class="rounded-md border border-dashed p-6 text-sm text-muted-foreground">
            {{ t('ledgers.empty') }}
          </div>
          <div v-else class="space-y-6">
            <div v-if="ownedLedgers.length" class="space-y-3">
              <p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                {{ t('ledgers.list.ownedTitle') }}
              </p>
              <div
                v-for="item in ownedLedgers"
                :key="item.id"
                class="flex flex-col gap-3 rounded-lg border border-border/60 p-4 md:flex-row md:items-center md:justify-between"
              >
                <div class="space-y-1">
                  <p class="font-medium">{{ item.name }}</p>
                  <div class="flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
                    <span>{{ item.currency_code }}</span>
                    <span
                      class="inline-flex items-center gap-1 rounded-full border px-2 py-0.5 text-xs"
                      :class="roleMeta(item.role).class"
                    >
                      <component :is="roleMeta(item.role).icon" class="h-3.5 w-3.5" />
                      {{ roleMeta(item.role).label }}
                    </span>
                  </div>
                </div>
                <Button
                  size="sm"
                  :variant="currentLedgerId === item.id ? 'default' : 'outline'"
                  @click="handleSelect(item.id)"
                >
                  {{ currentLedgerId === item.id ? t('ledgers.list.active') : t('ledgers.list.select') }}
                </Button>
              </div>
            </div>

            <div v-if="sharedLedgers.length" class="space-y-3">
              <p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                {{ t('ledgers.list.sharedTitle') }}
              </p>
              <div
                v-for="item in sharedLedgers"
                :key="item.id"
                class="flex flex-col gap-3 rounded-lg border border-border/60 p-4 md:flex-row md:items-center md:justify-between"
              >
                <div class="space-y-1">
                  <p class="font-medium">{{ item.name }}</p>
                  <div class="flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
                    <span>{{ item.currency_code }}</span>
                    <span
                      class="inline-flex items-center gap-1 rounded-full border px-2 py-0.5 text-xs"
                      :class="roleMeta(item.role).class"
                    >
                      <component :is="roleMeta(item.role).icon" class="h-3.5 w-3.5" />
                      {{ roleMeta(item.role).label }}
                    </span>
                  </div>
                </div>
                <Button
                  size="sm"
                  :variant="currentLedgerId === item.id ? 'default' : 'outline'"
                  @click="handleSelect(item.id)"
                >
                  {{ currentLedgerId === item.id ? t('ledgers.list.active') : t('ledgers.list.select') }}
                </Button>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>{{ t('ledgers.create.title') }}</CardTitle>
          <CardDescription>{{ t('ledgers.create.description') }}</CardDescription>
        </CardHeader>
        <CardContent class="space-y-4">
          <div class="space-y-2">
            <Label for="ledgerName">{{ t('ledgers.create.nameLabel') }}</Label>
            <Input
              id="ledgerName"
              ref="nameInput"
              v-model="form.name"
              :placeholder="defaultLedgerName || t('ledgers.create.namePlaceholder')"
            />
          </div>
          <div class="space-y-2">
            <Label for="ledgerCurrency">{{ t('ledgers.create.currencyLabel') }}</Label>
            <Select v-model="form.currency_code">
              <SelectTrigger id="ledgerCurrency">
                <SelectValue :placeholder="t('ledgers.create.currencyPlaceholder')" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="BRL">BRL</SelectItem>
                <SelectItem value="USD">USD</SelectItem>
                <SelectItem value="EUR">EUR</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <Button class="w-full" :disabled="isCreating" @click="handleCreate">
            {{ isCreating ? t('ledgers.create.submitting') : t('ledgers.create.submit') }}
          </Button>
        </CardContent>
      </Card>
    </div>
  </div>
</template>
