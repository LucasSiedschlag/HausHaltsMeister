<script setup lang="ts">
import { Button } from '@shared/components/ui/button'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@shared/components/ui/select'
import { useLedger } from '@shared/composables/useLedger'

const { t } = useI18n()
const { ledgers, currentLedgerId, fetchLedgers, isLoading } = useLedger()

const hasLedgers = computed(() => ledgers.value.length > 0)

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
</template>
