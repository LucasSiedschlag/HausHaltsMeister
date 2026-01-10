<script setup lang="ts">
import { computed } from 'vue'
import { Crown, Eye, Pencil } from 'lucide-vue-next'
import { Button } from '@shared/components/ui/button'
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectSeparator,
  SelectTrigger,
  SelectValue
} from '@shared/components/ui/select'
import { useLedger } from '@shared/composables/useLedger'
import type { LedgerRole } from '#layers/shared/utils/ledger-roles'

const { t } = useI18n()
const { ledgers, currentLedgerId, currentLedger, fetchLedgers, isLoading } = useLedger()

const hasLedgers = computed(() => ledgers.value.length > 0)
const ownedLedgers = computed(() => ledgers.value.filter((ledger) => ledger.role === 'owner'))
const sharedLedgers = computed(() => ledgers.value.filter((ledger) => ledger.role !== 'owner'))
const currentRoleMeta = computed(() => roleMeta(currentLedger.value?.role ?? null))

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

onMounted(async () => {
  if (!hasLedgers.value) {
    await fetchLedgers()
  }
})
</script>

<template>
  <div class="flex items-center gap-2">
    <Select v-if="hasLedgers" v-model="currentLedgerId" :disabled="isLoading">
      <SelectTrigger class="w-[220px]">
        <div v-if="currentLedger" class="flex w-full items-center gap-2">
          <span class="inline-flex h-5 w-5 items-center justify-center rounded-full border" :class="currentRoleMeta.class">
            <component :is="currentRoleMeta.icon" class="h-3 w-3" />
          </span>
          <span class="min-w-0 flex-1 truncate">{{ currentLedger.name }}</span>
        </div>
        <SelectValue v-else :placeholder="t('ledgers.selectPlaceholder')" />
      </SelectTrigger>
      <SelectContent>
        <SelectGroup v-if="ownedLedgers.length">
          <SelectLabel>{{ t('ledgers.list.ownedTitle') }}</SelectLabel>
          <SelectItem v-for="item in ownedLedgers" :key="item.id" :value="item.id">
            <div class="flex w-full items-center justify-between gap-3">
              <span class="truncate">{{ item.name }}</span>
              <span
                class="inline-flex items-center gap-1 rounded-full border px-2 py-0.5 text-[11px]"
                :class="roleMeta(item.role).class"
              >
                <component :is="roleMeta(item.role).icon" class="h-3 w-3" />
                {{ roleMeta(item.role).label }}
              </span>
            </div>
          </SelectItem>
        </SelectGroup>
        <SelectSeparator v-if="ownedLedgers.length && sharedLedgers.length" />
        <SelectGroup v-if="sharedLedgers.length">
          <SelectLabel>{{ t('ledgers.list.sharedTitle') }}</SelectLabel>
          <SelectItem v-for="item in sharedLedgers" :key="item.id" :value="item.id">
            <div class="flex w-full items-center justify-between gap-3">
              <span class="truncate">{{ item.name }}</span>
              <span
                class="inline-flex items-center gap-1 rounded-full border px-2 py-0.5 text-[11px]"
                :class="roleMeta(item.role).class"
              >
                <component :is="roleMeta(item.role).icon" class="h-3 w-3" />
                {{ roleMeta(item.role).label }}
              </span>
            </div>
          </SelectItem>
        </SelectGroup>
      </SelectContent>
    </Select>
    <Button v-else variant="outline" size="sm" @click="navigateTo('/ledgers')">
      {{ t('ledgers.selectAction') }}
    </Button>
  </div>
</template>
