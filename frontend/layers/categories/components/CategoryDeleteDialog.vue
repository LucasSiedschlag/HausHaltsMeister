<script setup lang="ts">
import { ref, watch } from 'vue'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '~/layers/shared/components/ui/dialog'
import { Button } from '~/layers/shared/components/ui/button'
import type { Category } from '../types/category'

interface Props {
  open: boolean
  category?: Category | null
  submitting?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  category: null,
  submitting: false,
})

const step = ref<'choice' | 'confirm-current'>('choice')

const emit = defineEmits<{
  'update:open': [value: boolean]
  confirm: [scope: 'current' | 'next']
}>()

watch(
  () => props.open,
  (open) => {
    if (!open) step.value = 'choice'
  },
)

watch(
  () => props.category?.id,
  () => {
    step.value = 'choice'
  },
)
</script>

<template>
  <Dialog :open="props.open" @update:open="emit('update:open', $event)">
    <DialogContent class="sm:max-w-[580px]">
      <DialogHeader>
        <DialogTitle>Desativar categoria</DialogTitle>
        <DialogDescription v-if="step === 'choice'">
          Escolha quando a categoria
          <span class="font-medium text-foreground">{{ props.category?.name }}</span>
          deve ser desativada.
        </DialogDescription>
        <DialogDescription v-else>
          Há registros vinculados à categoria
          <span class="font-medium text-foreground">{{ props.category?.name }}</span>
          no mês atual. Desativar agora pode impactar o planejamento do período. Deseja continuar?
        </DialogDescription>
      </DialogHeader>

      <DialogFooter v-if="step === 'choice'" class="mt-6 flex flex-col-reverse gap-2 sm:flex-row sm:justify-center">
        <Button variant="outline" :disabled="props.submitting" @click="emit('update:open', false)">
          Cancelar
        </Button>
        <Button variant="secondary" :disabled="props.submitting" @click="step = 'confirm-current'">
          Desativar neste mês
        </Button>
        <Button variant="destructive" :disabled="props.submitting" @click="emit('confirm', 'next')">
          Desativar a partir do próximo mês
        </Button>
      </DialogFooter>

      <DialogFooter v-else class="mt-6 flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
        <Button variant="outline" :disabled="props.submitting" @click="step = 'choice'">
          Voltar
        </Button>
        <Button variant="destructive" :disabled="props.submitting" @click="emit('confirm', 'current')">
          Desativar mesmo assim
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
