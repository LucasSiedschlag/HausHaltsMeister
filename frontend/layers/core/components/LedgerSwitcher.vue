<script setup lang="ts">
import { Button } from '@shared/components/ui/button'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@shared/components/ui/select'
import { useLedger } from '@shared/composables/useLedger'
import { useLedgerContext } from '@shared/composables/useLedgerContext'

const { t } = useI18n()
const { ledgers, currentLedgerId, fetchLedgers, isLoading } = useLedger()
const ledgerContext = useLedgerContext()

const hasLedgers = computed(() => ledgers.value.length > 0)
const roleBadge = computed(() => {
  const role = ledgerContext.activeRole.value
  if (!role) return ''
  const roleLabel = t(`ledgers.roles.${role}`)
  return t('ledgers.roleBadge', { role: roleLabel })
})

onMounted(async () => {
  if (!hasLedgers.value) {
    await fetchLedgers()
  }
})
</script>

<template>
  <div class="flex flex-col gap-1">
    <div class="flex items-center gap-2">
      <Select v-if="hasLedgers" v-model="currentLedgerId" :disabled="isLoading">
        <SelectTrigger class="w-[220px]">
          <SelectValue :placeholder="t('ledgers.selectPlaceholder')" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem v-for="item in ledgers" :key="item.id" :value="item.id">
            {{ item.name }}
          </SelectItem>
        </SelectContent>
      </Select>
      <Button v-else variant="outline" size="sm" @click="navigateTo('/ledgers')">
        {{ t('ledgers.selectAction') }}
      </Button>
    </div>
    <p v-if="roleBadge" class="text-[11px] text-muted-foreground">{{ roleBadge }}</p>
  </div>
</template>
