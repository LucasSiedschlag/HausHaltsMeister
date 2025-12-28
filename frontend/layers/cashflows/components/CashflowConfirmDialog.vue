<script setup lang="ts">
import { computed } from 'vue'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '~/layers/shared/components/ui/dialog'
import { Button } from '~/layers/shared/components/ui/button'
import type { CashflowEntry } from '../types/cashflow'

interface Props {
  open: boolean
  mode: 'delete' | 'reverse'
  entry?: CashflowEntry | null
  submitting?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  entry: null,
  submitting: false,
})

const emit = defineEmits<{
  'update:open': [value: boolean]
  confirm: []
}>()

const dialogTitle = computed(() => (props.mode === 'delete' ? 'Excluir lançamento' : 'Estornar lançamento'))

const dialogDescription = computed(() => {
  if (props.mode === 'delete') {
    return 'Esta ação remove o lançamento definitivamente.'
  }
  return 'O estorno cria um lançamento inverso para ajustar o saldo.'
})

const confirmLabel = computed(() => (props.mode === 'delete' ? 'Excluir' : 'Estornar'))

const confirmVariant = computed(() => (props.mode === 'delete' ? 'destructive' : 'secondary'))
</script>

<template>
  <Dialog :open="props.open" @update:open="emit('update:open', $event)">
    <DialogContent class="sm:max-w-[520px]">
      <DialogHeader>
        <DialogTitle>{{ dialogTitle }}</DialogTitle>
        <DialogDescription>
          {{ dialogDescription }}
          <span v-if="props.entry" class="font-medium text-foreground">
            {{ props.entry.title }}
          </span>
        </DialogDescription>
      </DialogHeader>

      <DialogFooter class="mt-6 flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
        <Button variant="outline" :disabled="props.submitting" @click="emit('update:open', false)">
          Cancelar
        </Button>
        <Button :variant="confirmVariant" :disabled="props.submitting" @click="emit('confirm')">
          {{ confirmLabel }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
