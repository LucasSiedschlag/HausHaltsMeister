<script setup lang="ts">
import { ref } from 'vue'
import { Button } from '@shared/components/ui/button'
import { Card, CardContent } from '@shared/components/ui/card'
import { Input } from '@shared/components/ui/input'
import { Label } from '@shared/components/ui/label'
import { Separator } from '@shared/components/ui/separator'
import googleIcon from '@shared/icons/google.svg'
import githubIcon from '@shared/icons/github.svg'
import { useAuth } from '#layers/auth/composables/useAuth'
import { emailSchema, passwordSchema, useInlineValidation, getInputClass } from '@shared/validators'

const { t } = useI18n()

const email = ref('')
const password = ref('')
const remember = ref(false)
const oauthLoading = ref<'google' | 'github' | null>(null)
const error = ref('')
const isLoading = ref(false)

const { login, startOAuth, extractErrorMessage, extractErrorDetails, extractErrorCode } = useAuth()
const router = useRouter()

const { touched, errors, touchField, validateAll } = useInlineValidation(
  { email, password },
  { email: emailSchema(), password: passwordSchema() }
)

const onSubmit = async () => {
  error.value = ''
  if (validateAll()) {
    return
  }

  isLoading.value = true
  try {
    await login(email.value, password.value, remember.value)
    await router.push('/')
  } catch (err) {
    const code = extractErrorCode(err)
    if (code === 'AUTH_INVALID_CREDENTIALS') {
      error.value = t('auth.login.errors.invalidCredentials')
      return
    }
    if (code === 'VALIDATION_ERROR') {
      const details = extractErrorDetails(err)
      if (details.email === 'invalid') {
        error.value = t('auth.login.errors.invalidEmail')
        return
      }
      if (details.password === 'min_length') {
        error.value = t('auth.login.errors.passwordMin')
        return
      }
    }
    error.value = extractErrorMessage(err)
  } finally {
    isLoading.value = false
  }
}

const handleOAuth = (provider: 'google' | 'github') => {
  oauthLoading.value = provider
  startOAuth(provider)
}
</script>

<template>
  <div class="flex flex-col gap-6">
    <Card class="overflow-hidden p-0">
      <CardContent class="grid p-0 md:grid-cols-2">
        <form class="p-6 md:p-8" @submit.prevent="onSubmit">
            <div class="flex flex-col gap-6">
              <div class="flex flex-col items-center gap-2 text-center">
                <h1 class="text-2xl font-bold">{{ t('auth.login.title') }}</h1>
                <p class="text-balance text-sm text-muted-foreground">
                  {{ t('auth.login.subtitle') }}
                </p>
              </div>
              <div class="grid gap-2">
                <Label for-id="email">{{ t('auth.login.emailLabel') }} <span class="text-destructive">*</span></Label>
              <Input
                id="email"
                v-model="email"
                type="email"
                :placeholder="t('auth.login.emailPlaceholder')"
                :class="getInputClass({ touched: touched.email, hasError: Boolean(errors.email[0]) })"
                @blur="touchField('email')"
              />
              <p v-if="touched.email && errors.email[0]" class="text-xs text-destructive">
                {{ errors.email[0].message }}
              </p>
            </div>
            <div class="grid gap-2">
              <div class="flex items-center">
                <Label for-id="password">{{ t('auth.login.passwordLabel') }} <span class="text-destructive">*</span></Label>
                <NuxtLink to="/auth/forgot-password" class="ml-auto text-sm underline-offset-2 hover:underline">
                  {{ t('auth.login.forgotPassword') }}
                </NuxtLink>
              </div>
              <Input
                id="password"
                v-model="password"
                type="password"
                :class="getInputClass({ touched: touched.password, hasError: Boolean(errors.password[0]) })"
                @blur="touchField('password')"
              />
              <p v-if="touched.password && errors.password[0]" class="text-xs text-destructive">
                {{ errors.password[0].message }}
              </p>
            </div>
            <div class="flex items-center gap-2 text-sm text-muted-foreground">
              <input
                id="remember"
                v-model="remember"
                type="checkbox"
                class="h-4 w-4 rounded border border-input text-primary focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/20"
              />
              <Label for-id="remember" class="text-sm text-muted-foreground">{{ t('auth.login.remember') }}</Label>
            </div>
            <Button type="submit" class="w-full" :disabled="isLoading">
              {{ isLoading ? t('auth.login.submitting') : t('auth.login.submit') }}
            </Button>
            <p v-if="error" class="text-center text-sm text-destructive">
              {{ error }}
            </p>
            <Separator>{{ t('auth.login.oauthSeparator') }}</Separator>
            <div class="grid grid-cols-2 gap-4">
              <Button variant="outline" type="button" :disabled="oauthLoading !== null" @click="handleOAuth('google')">
                <span
                  v-if="oauthLoading === 'google'"
                  class="h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent"
                />
                <img v-else :src="googleIcon" alt="" class="h-4 w-4" />
                <span>{{ t('auth.login.oauthGoogle') }}</span>
              </Button>
              <Button variant="outline" type="button" :disabled="oauthLoading !== null" @click="handleOAuth('github')">
                <span
                  v-if="oauthLoading === 'github'"
                  class="h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent"
                />
                <img v-else :src="githubIcon" alt="" class="h-4 w-4" />
                <span>{{ t('auth.login.oauthGithub') }}</span>
              </Button>
            </div>
            <div class="text-center text-sm text-muted-foreground">
              {{ t('auth.login.noAccount') }}
              <NuxtLink to="/auth/signup" class="underline underline-offset-4">{{ t('auth.login.signupLink') }}</NuxtLink>
            </div>
          </div>
        </form>
        <div class="bg-muted relative hidden md:block">
          <img
            src="/placeholder.svg"
            alt="Image"
            class="absolute inset-0 h-full w-full object-cover dark:brightness-[0.2] dark:grayscale"
          />
        </div>
      </CardContent>
    </Card>
    <p class="px-6 text-center text-xs text-muted-foreground">
      {{ t('auth.login.termsPrefix') }}
      <a href="#" class="underline underline-offset-4">{{ t('auth.login.termsLink') }}</a>
      {{ t('auth.login.termsMiddle') }}
      <a href="#" class="underline underline-offset-4">{{ t('auth.login.privacyLink') }}</a>.
    </p>
  </div>
</template>
