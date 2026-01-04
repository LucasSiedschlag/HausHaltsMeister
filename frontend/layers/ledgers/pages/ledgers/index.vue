<script setup lang="ts">
import { Button } from '@shared/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@shared/components/ui/card'
import { Input } from '@shared/components/ui/input'
import { Label } from '@shared/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@shared/components/ui/select'
import { useLedger } from '@shared/composables/useLedger'
import { push } from 'notivue'

const { t } = useI18n()
const router = useRouter()
const { ledgers, currentLedgerId, fetchLedgers, selectLedger, createLedger, isLoading } = useLedger()

const form = reactive({
  name: '',
  currency_code: 'BRL'
})
const isCreating = ref(false)

const loadLedgers = async () => {
  await fetchLedgers()
}

const handleSelect = async (ledgerId: string) => {
  selectLedger(ledgerId)
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

onMounted(async () => {
  await loadLedgers()
})
</script>

<template>
  <div class="space-y-6">
    <section class="space-y-2">
      <h1 class="text-2xl font-semibold">{{ t('ledgers.title') }}</h1>
      <p class="text-sm text-muted-foreground">{{ t('ledgers.description') }}</p>
    </section>

    <div class="grid gap-4 lg:grid-cols-[1.4fr_1fr]">
      <Card>
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
          <div v-else class="space-y-3">
            <div
              v-for="item in ledgers"
              :key="item.id"
              class="flex flex-col gap-3 rounded-lg border border-border/60 p-4 md:flex-row md:items-center md:justify-between"
            >
              <div class="space-y-1">
                <p class="font-medium">{{ item.name }}</p>
                <p class="text-xs text-muted-foreground">
                  {{ t('ledgers.list.role') }}: {{ item.role }} · {{ item.currency_code }}
                </p>
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
            <Input id="ledgerName" v-model="form.name" :placeholder="t('ledgers.create.namePlaceholder')" />
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
