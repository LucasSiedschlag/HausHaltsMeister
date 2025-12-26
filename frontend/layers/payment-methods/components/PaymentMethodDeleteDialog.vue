<script setup lang="ts">
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '~/layers/shared/components/ui/dialog'
import { Button } from '~/layers/shared/components/ui/button'
import type { PaymentMethod } from '../types/payment-method'

interface Props {
  open: boolean
  method?: PaymentMethod | null
  submitting?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  method: null,
  submitting: false,
})

const emit = defineEmits<{
  'update:open': [value: boolean]
  confirm: []
}>()
</script>

<template>
  <Dialog :open="props.open" @update:open="emit('update:open', $event)">
    <DialogContent>
      <DialogHeader>
        <DialogTitle>Desativar meio de pagamento</DialogTitle>
        <DialogDescription>
          Tem certeza que deseja desativar o meio de pagamento
          <span class="font-medium text-foreground">{{ props.method?.name }}</span>?
          Ele deixará de aparecer nas seleções futuras.
        </DialogDescription>
      </DialogHeader>
      <DialogFooter class="mt-6">
        <Button variant="outline" :disabled="props.submitting" @click="emit('update:open', false)">
          Cancelar
        </Button>
        <Button variant="destructive" :disabled="props.submitting" @click="emit('confirm')">
          Desativar
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
