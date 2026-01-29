<template>
  <div
    class="relative flex min-h-screen flex-col overflow-hidden bg-gradient-to-br from-gray-50 via-primary-50/30 to-gray-100 dark:from-dark-950 dark:via-dark-900 dark:to-dark-950"
  >
    <!-- Background Decorations -->
    <div class="pointer-events-none absolute inset-0 overflow-hidden">
      <div
        class="absolute -right-40 -top-40 h-96 w-96 rounded-full bg-primary-400/20 blur-3xl"
      ></div>
      <div
        class="absolute -bottom-40 -left-40 h-96 w-96 rounded-full bg-primary-500/15 blur-3xl"
      ></div>
      <div
        class="absolute left-1/3 top-1/4 h-72 w-72 rounded-full bg-primary-300/10 blur-3xl"
      ></div>
    </div>

    <!-- Header -->
    <header class="relative z-20 px-6 py-4">
      <nav class="mx-auto flex max-w-6xl items-center justify-between">
        <!-- Logo -->
        <div class="flex items-center">
          <router-link to="/home" class="flex items-center gap-3">
            <div class="h-10 w-10 overflow-hidden rounded-xl shadow-md">
              <img :src="siteLogo || '/logo.png'" alt="Logo" class="h-full w-full object-contain" />
            </div>
            <span class="text-xl font-bold text-gray-900 dark:text-white">{{ siteName }}</span>
          </router-link>
        </div>

        <!-- Nav Actions -->
        <div class="flex items-center gap-3">
          <!-- Theme Toggle -->
          <button
            @click="toggleTheme"
            class="rounded-lg p-2 text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-700 dark:text-dark-400 dark:hover:bg-dark-800 dark:hover:text-white"
            :title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
          >
            <Icon v-if="isDark" name="sun" size="md" />
            <Icon v-else name="moon" size="md" />
          </button>

          <!-- Login / Dashboard Button -->
          <router-link
            v-if="isAuthenticated"
            :to="dashboardPath"
            class="inline-flex items-center gap-1.5 rounded-full bg-gray-900 py-1 pl-1 pr-2.5 transition-colors hover:bg-gray-800 dark:bg-gray-800 dark:hover:bg-gray-700"
          >
            <span
              class="flex h-5 w-5 items-center justify-center rounded-full bg-gradient-to-br from-primary-400 to-primary-600 text-[10px] font-semibold text-white"
            >
              {{ userInitial }}
            </span>
            <span class="text-xs font-medium text-white">{{ t('home.dashboard') }}</span>
          </router-link>
          <router-link
            v-else
            to="/login"
            class="inline-flex items-center rounded-full bg-gray-900 px-3 py-1 text-xs font-medium text-white transition-colors hover:bg-gray-800 dark:bg-gray-800 dark:hover:bg-gray-700"
          >
            {{ t('home.login') }}
          </router-link>
        </div>
      </nav>
    </header>

    <!-- Main Content -->
    <main class="relative z-10 flex-1 px-6 py-8">
      <div class="mx-auto max-w-6xl">
        <!-- Page Title -->
        <div class="mb-12 text-center">
          <h1 class="mb-4 text-4xl font-bold text-gray-900 dark:text-white md:text-5xl">
            {{ t('guide.title') }}
          </h1>
          <p class="text-lg text-gray-600 dark:text-dark-300">
            {{ t('guide.subtitle') }}
          </p>
        </div>

        <!-- Tool Tabs -->
        <div class="mb-8 flex justify-center">
          <div class="inline-flex rounded-xl bg-white/80 p-1.5 shadow-lg backdrop-blur-sm dark:bg-dark-800/80">
            <button
              v-for="tool in tools"
              :key="tool.id"
              @click="activeTool = tool.id"
              :class="[
                'rounded-lg px-6 py-3 text-sm font-medium transition-all',
                activeTool === tool.id
                  ? `${tool.activeClass} text-white shadow-md`
                  : 'text-gray-600 hover:bg-gray-100 dark:text-dark-300 dark:hover:bg-dark-700'
              ]"
            >
              {{ tool.name }}
            </button>
          </div>
        </div>

        <!-- Claude Code Section -->
        <div v-show="activeTool === 'claude-code'" class="space-y-8">
          <!-- Quick Start -->
          <div class="rounded-2xl bg-gradient-to-r from-blue-50 to-indigo-50 p-8 text-center dark:from-blue-900/20 dark:to-indigo-900/20">
            <h3 class="mb-4 text-2xl font-bold text-blue-900 dark:text-blue-100">🚀 Claude Code {{ t('guide.quickStart') }}</h3>
            <p class="mb-6 text-blue-700 dark:text-blue-200">Anthropic {{ t('guide.officialCli') }}，Claude Sonnet 4.5 {{ t('guide.powered') }}</p>
            <div class="flex flex-wrap items-center justify-center gap-8 text-blue-600 dark:text-blue-300">
              <div class="flex items-center">
                <span class="mr-3 flex h-8 w-8 items-center justify-center rounded-full bg-blue-600 font-bold text-white">1</span>
                <span>{{ t('guide.installCli') }}</span>
              </div>
              <div class="flex items-center">
                <span class="mr-3 flex h-8 w-8 items-center justify-center rounded-full bg-blue-600 font-bold text-white">2</span>
                <span>{{ t('guide.configureKey') }}</span>
              </div>
              <div class="flex items-center">
                <span class="mr-3 flex h-8 w-8 items-center justify-center rounded-full bg-blue-600 font-bold text-white">3</span>
                <span>{{ t('guide.startCoding') }}</span>
              </div>
            </div>
          </div>

          <!-- Platform Tabs -->
          <div class="flex justify-center">
            <div class="inline-flex rounded-xl bg-white p-1 shadow-lg dark:bg-dark-800">
              <button
                v-for="platform in platforms"
                :key="platform.id"
                @click="activePlatform = platform.id"
                :class="[
                  'rounded-lg px-6 py-3 text-sm font-medium transition-all',
                  activePlatform === platform.id
                    ? 'bg-blue-600 text-white shadow-md'
                    : 'text-gray-600 hover:bg-gray-100 dark:text-dark-300 dark:hover:bg-dark-700'
                ]"
              >
                {{ platform.icon }} {{ platform.name }}
              </button>
            </div>
          </div>

          <!-- Installation Steps -->
          <div class="rounded-2xl bg-white p-8 shadow-lg dark:bg-dark-800">
            <h3 class="mb-6 flex items-center text-2xl font-bold text-gray-900 dark:text-white">
              <span class="mr-4 flex h-10 w-10 items-center justify-center rounded-full bg-blue-600 text-white">
                {{ platformIcons[activePlatform] }}
              </span>
              {{ platformNames[activePlatform] }} {{ t('guide.tutorial') }}
            </h3>

            <!-- Step 1: Install Node.js -->
            <div class="mb-8">
              <h4 class="mb-4 flex items-center text-xl font-semibold text-gray-800 dark:text-gray-200">
                <span class="mr-3 flex h-8 w-8 items-center justify-center rounded-full bg-blue-600 text-sm font-bold text-white">1</span>
                {{ t('guide.installNodejs') }}
              </h4>
              <div class="rounded-lg bg-gray-50 p-6 dark:bg-dark-700">
                <p class="mb-4 text-gray-700 dark:text-gray-300">
                  <strong>{{ t('guide.method1') }}{{ t('guide.officialInstaller') }}{{ t('guide.recommended') }}</strong>
                </p>
                <ol class="mb-4 ml-4 list-inside list-decimal space-y-2 text-gray-700 dark:text-gray-300">
                  <li>{{ t('guide.visitNodejs') }} <a href="https://nodejs.org/" target="_blank" class="text-blue-600 hover:underline">https://nodejs.org</a></li>
                  <li>{{ t('guide.downloadLts') }}</li>
                  <li>{{ t('guide.runInstaller') }}</li>
                </ol>

                <template v-if="activePlatform === 'windows'">
                  <p class="mb-3 mt-6 text-gray-700 dark:text-gray-300"><strong>{{ t('guide.method2') }}{{ t('guide.packageManager') }}</strong></p>
                  <CodeBlock :code="'winget install OpenJS.NodeJS.LTS'" :title="'PowerShell'" />
                </template>

                <template v-if="activePlatform === 'macos'">
                  <p class="mb-3 mt-6 text-gray-700 dark:text-gray-300"><strong>{{ t('guide.method2') }}Homebrew{{ t('guide.recommended') }}</strong></p>
                  <CodeBlock :code="'brew install node'" :title="'Terminal'" />
                </template>

                <template v-if="activePlatform === 'linux'">
                  <p class="mb-3 mt-6 text-gray-700 dark:text-gray-300"><strong>{{ t('guide.method2') }}{{ t('guide.packageManager') }}</strong></p>
                  <CodeBlock :code="'# Ubuntu/Debian\nsudo apt update && sudo apt install -y nodejs npm\n\n# CentOS/RHEL\nsudo yum install -y nodejs npm'" :title="'Terminal'" />
                </template>

                <p class="mb-3 mt-6 text-gray-700 dark:text-gray-300"><strong>{{ t('guide.verifyInstall') }}</strong></p>
                <CodeBlock :code="'node --version\nnpm --version'" :title="activePlatform === 'windows' ? 'CMD/PowerShell' : 'Terminal'" />
              </div>
            </div>

            <!-- Step 2: Install Claude Code CLI -->
            <div class="mb-8">
              <h4 class="mb-4 flex items-center text-xl font-semibold text-gray-800 dark:text-gray-200">
                <span class="mr-3 flex h-8 w-8 items-center justify-center rounded-full bg-blue-600 text-sm font-bold text-white">2</span>
                {{ t('guide.installClaudeCode') }}
              </h4>
              <div class="rounded-lg bg-gray-50 p-6 dark:bg-dark-700">
                <CodeBlock :code="'npm install -g @anthropic-ai/claude-code'" :title="activePlatform === 'windows' ? 'CMD/PowerShell' : 'Terminal'" />
                <p class="mb-3 mt-4 text-gray-700 dark:text-gray-300"><strong>{{ t('guide.verifyInstall') }}</strong></p>
                <CodeBlock :code="'claude --version'" :title="activePlatform === 'windows' ? 'CMD/PowerShell' : 'Terminal'" />
              </div>
            </div>

            <!-- Step 3: Configure API -->
            <div class="mb-8">
              <h4 class="mb-4 flex items-center text-xl font-semibold text-gray-800 dark:text-gray-200">
                <span class="mr-3 flex h-8 w-8 items-center justify-center rounded-full bg-blue-600 text-sm font-bold text-white">3</span>
                {{ t('guide.configureApi') }}
              </h4>
              <div class="rounded-lg bg-gray-50 p-6 dark:bg-dark-700">
                <div class="mb-6">
                  <h5 class="mb-3 text-lg font-semibold text-gray-800 dark:text-gray-200">3.1 {{ t('guide.getAuthToken') }}</h5>
                  <p class="mb-3 text-gray-700 dark:text-gray-300">
                    {{ t('guide.visitConsole') }} <a :href="tokenUrl" target="_blank" class="font-medium text-blue-600 hover:text-blue-800 dark:text-blue-400">{{ t('guide.console') }}</a> {{ t('guide.andDo') }}
                  </p>
                  <ul class="mb-4 ml-4 space-y-2 text-gray-700 dark:text-gray-300">
                    <li>• {{ t('guide.clickAddToken') }}</li>
                    <li>• <strong>{{ t('guide.selectGroup') }}Claude Code{{ t('guide.dedicated') }}{{ t('guide.mustSelect') }}</strong></li>
                    <li>• {{ t('guide.tokenName') }}</li>
                    <li>• {{ t('guide.quotaSuggestion') }}</li>
                  </ul>
                </div>

                <div>
                  <h5 class="mb-3 text-lg font-semibold text-gray-800 dark:text-gray-200">3.2 {{ t('guide.configureEnv') }}</h5>
                  <div class="mb-4 rounded-r-lg border-l-4 border-red-400 bg-red-50 p-4 dark:bg-red-900/20">
                    <p class="font-medium text-red-700 dark:text-red-300">
                      {{ t('guide.importantNote') }}{{ t('guide.replaceToken') }}
                    </p>
                  </div>

                  <p class="mb-3 text-gray-700 dark:text-gray-300">
                    {{ t('guide.configLocation') }}<code class="rounded bg-gray-200 px-2 py-1 text-sm text-gray-800 dark:bg-dark-600 dark:text-gray-200">{{ activePlatform === 'windows' ? '%USERPROFILE%\\.claude\\settings.json' : '~/.claude/settings.json' }}</code>
                  </p>
                  <CodeBlock
                    :code='`{\n  "env": {\n    "ANTHROPIC_AUTH_TOKEN": "` + t("guide.pasteYourToken") + `",\n    "ANTHROPIC_BASE_URL": "` + apiBaseUrl + `"\n  }\n}`'
                    title="settings.json"
                    :copyText="t('guide.copyConfig')"
                  />
                </div>
              </div>
            </div>

            <!-- Step 4: Start Claude Code -->
            <div>
              <h4 class="mb-4 flex items-center text-xl font-semibold text-gray-800 dark:text-gray-200">
                <span class="mr-3 flex h-8 w-8 items-center justify-center rounded-full bg-blue-600 text-sm font-bold text-white">4</span>
                {{ t('guide.startClaudeCode') }}
              </h4>
              <div class="rounded-lg bg-gray-50 p-6 dark:bg-dark-700">
                <p class="mb-4 text-gray-700 dark:text-gray-300">{{ t('guide.enterProjectDir') }}</p>
                <CodeBlock :code="'cd your-project-folder'" :title="activePlatform === 'windows' ? 'CMD/PowerShell' : 'Terminal'" />
                <p class="my-4 text-gray-700 dark:text-gray-300">{{ t('guide.thenRun') }}</p>
                <CodeBlock :code="'claude'" :title="activePlatform === 'windows' ? 'CMD/PowerShell' : 'Terminal'" />
                <div class="mt-4 rounded-r-lg border-l-4 border-green-400 bg-green-50 p-4 dark:bg-green-900/20">
                  <p class="mb-2 font-semibold text-green-700 dark:text-green-300">{{ t('guide.firstStartNote') }}</p>
                  <ul class="space-y-1 text-green-700 dark:text-green-300">
                    <li>• {{ t('guide.selectTheme') }}</li>
                    <li>• {{ t('guide.confirmSecurity') }}</li>
                    <li>• {{ t('guide.useDefaultTerminal') }}</li>
                    <li>• {{ t('guide.trustWorkDir') }}</li>
                    <li>• {{ t('guide.startCodingNote') }}</li>
                  </ul>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- CodeX Section -->
        <div v-show="activeTool === 'codex'" class="space-y-8">
          <!-- Quick Start -->
          <div class="rounded-2xl bg-gradient-to-r from-purple-50 to-pink-50 p-8 text-center dark:from-purple-900/20 dark:to-pink-900/20">
            <h3 class="mb-4 text-2xl font-bold text-purple-900 dark:text-purple-100">🚀 CodeX {{ t('guide.quickStart') }}</h3>
            <p class="mb-6 text-purple-700 dark:text-purple-200">{{ t('guide.enterpriseAi') }}，GPT-5 {{ t('guide.powered') }}</p>
            <div class="flex flex-wrap items-center justify-center gap-8 text-purple-600 dark:text-purple-300">
              <div class="flex items-center">
                <span class="mr-3 flex h-8 w-8 items-center justify-center rounded-full bg-purple-600 font-bold text-white">1</span>
                <span>{{ t('guide.envPrep') }}</span>
              </div>
              <div class="flex items-center">
                <span class="mr-3 flex h-8 w-8 items-center justify-center rounded-full bg-purple-600 font-bold text-white">2</span>
                <span>{{ t('guide.installConfig') }}</span>
              </div>
              <div class="flex items-center">
                <span class="mr-3 flex h-8 w-8 items-center justify-center rounded-full bg-purple-600 font-bold text-white">3</span>
                <span>{{ t('guide.startCoding') }}</span>
              </div>
            </div>
          </div>

          <!-- Platform Tabs -->
          <div class="flex justify-center">
            <div class="inline-flex rounded-xl bg-white p-1 shadow-lg dark:bg-dark-800">
              <button
                v-for="platform in platforms"
                :key="platform.id"
                @click="activePlatform = platform.id"
                :class="[
                  'rounded-lg px-6 py-3 text-sm font-medium transition-all',
                  activePlatform === platform.id
                    ? 'bg-purple-600 text-white shadow-md'
                    : 'text-gray-600 hover:bg-gray-100 dark:text-dark-300 dark:hover:bg-dark-700'
                ]"
              >
                {{ platform.icon }} {{ platform.name }}
              </button>
            </div>
          </div>

          <!-- Installation Steps -->
          <div class="rounded-2xl bg-white p-8 shadow-lg dark:bg-dark-800">
            <h3 class="mb-6 flex items-center text-2xl font-bold text-gray-900 dark:text-white">
              <span class="mr-4 flex h-10 w-10 items-center justify-center rounded-full bg-purple-600 text-white">
                {{ platformIcons[activePlatform] }}
              </span>
              {{ platformNames[activePlatform] }} {{ t('guide.tutorial') }}
            </h3>

            <!-- Step 1: Install Node.js -->
            <div class="mb-8">
              <h4 class="mb-4 flex items-center text-xl font-semibold text-gray-800 dark:text-gray-200">
                <span class="mr-3 flex h-8 w-8 items-center justify-center rounded-full bg-purple-600 text-sm font-bold text-white">1</span>
                {{ t('guide.installNodejs') }}
              </h4>
              <div class="rounded-lg bg-gray-50 p-6 dark:bg-dark-700">
                <p class="mb-4 text-gray-700 dark:text-gray-300">{{ t('guide.sameAsClaudeCode') }}</p>
                <CodeBlock :code="'node --version\nnpm --version'" :title="activePlatform === 'windows' ? 'CMD/PowerShell' : 'Terminal'" />
              </div>
            </div>

            <!-- Step 2: Install CodeX CLI -->
            <div class="mb-8">
              <h4 class="mb-4 flex items-center text-xl font-semibold text-gray-800 dark:text-gray-200">
                <span class="mr-3 flex h-8 w-8 items-center justify-center rounded-full bg-purple-600 text-sm font-bold text-white">2</span>
                {{ t('guide.installCodex') }}
              </h4>
              <div class="rounded-lg bg-gray-50 p-6 dark:bg-dark-700">
                <CodeBlock :code="'npm install -g @openai/codex@latest'" :title="activePlatform === 'windows' ? 'CMD/PowerShell' : 'Terminal'" />
                <p class="mb-3 mt-4 text-gray-700 dark:text-gray-300"><strong>{{ t('guide.verifyInstall') }}</strong></p>
                <CodeBlock :code="'codex --version'" :title="activePlatform === 'windows' ? 'CMD/PowerShell' : 'Terminal'" />
              </div>
            </div>

            <!-- Step 3: Configure API -->
            <div class="mb-8">
              <h4 class="mb-4 flex items-center text-xl font-semibold text-gray-800 dark:text-gray-200">
                <span class="mr-3 flex h-8 w-8 items-center justify-center rounded-full bg-purple-600 text-sm font-bold text-white">3</span>
                {{ t('guide.configureApi') }}
              </h4>
              <div class="rounded-lg bg-gray-50 p-6 dark:bg-dark-700">
                <div class="mb-6">
                  <h5 class="mb-3 text-lg font-semibold text-gray-800 dark:text-gray-200">3.1 {{ t('guide.getCodexToken') }}</h5>
                  <ul class="mb-4 ml-4 space-y-2 text-gray-700 dark:text-gray-300">
                    <li>• {{ t('guide.visitConsole') }} <a :href="tokenUrl" target="_blank" class="font-medium text-blue-600 hover:text-blue-800 dark:text-blue-400">{{ t('guide.console') }}</a></li>
                    <li>• {{ t('guide.clickAddToken') }}</li>
                    <li>• <strong>{{ t('guide.selectGroup') }}CodeX{{ t('guide.dedicated') }}{{ t('guide.mustSelect') }}</strong></li>
                  </ul>
                  <div class="mb-4 rounded border-l-4 border-red-400 bg-red-50 p-4 dark:bg-red-900/20">
                    <p class="text-sm font-medium text-red-800 dark:text-red-300">
                      <strong>{{ t('guide.important') }}</strong>{{ t('guide.codexDifferentToken') }}
                    </p>
                  </div>
                </div>

                <div class="mb-6">
                  <h5 class="mb-3 text-lg font-semibold text-gray-800 dark:text-gray-200">3.2 {{ t('guide.createConfigDir') }}</h5>
                  <CodeBlock
                    :code="activePlatform === 'windows' ? 'mkdir %USERPROFILE%\\.codex\ncd %USERPROFILE%\\.codex' : 'mkdir -p ~/.codex\ncd ~/.codex'"
                    :title="activePlatform === 'windows' ? 'CMD/PowerShell' : 'Terminal'"
                  />
                </div>

                <div class="mb-6">
                  <h5 class="mb-3 text-lg font-semibold text-gray-800 dark:text-gray-200">3.3 {{ t('guide.createConfigFile') }}</h5>
                  <p class="mb-3 text-gray-700 dark:text-gray-300">{{ t('guide.createConfigToml') }}</p>
                  <CodeBlock
                    :code='`model_provider = "duckcoding"\nmodel = "gpt-5.2-codex"\nmodel_reasoning_effort = "xhigh"\nnetwork_access = "enabled"\ndisable_response_storage = true\n\n[model_providers.duckcoding]\nname = "duckcoding"\nbase_url = "` + apiBaseUrl + `/v1"\nwire_api = "responses"\nrequires_openai_auth = true`'
                    title="config.toml"
                    :copyText="t('guide.copyConfig')"
                  />

                  <p class="mb-3 mt-4 text-gray-700 dark:text-gray-300">{{ t('guide.createAuthJson') }}</p>
                  <CodeBlock
                    :code='`{\n  "OPENAI_API_KEY": "` + t("guide.pasteCodexToken") + `"\n}`'
                    title="auth.json"
                    :copyText="t('guide.copyConfig')"
                  />
                </div>
              </div>
            </div>

            <!-- Step 4: Start CodeX -->
            <div>
              <h4 class="mb-4 flex items-center text-xl font-semibold text-gray-800 dark:text-gray-200">
                <span class="mr-3 flex h-8 w-8 items-center justify-center rounded-full bg-purple-600 text-sm font-bold text-white">4</span>
                {{ t('guide.startCodex') }}
              </h4>
              <div class="rounded-lg bg-gray-50 p-6 dark:bg-dark-700">
                <CodeBlock :code="'cd your-project-folder\ncodex'" :title="activePlatform === 'windows' ? 'CMD/PowerShell' : 'Terminal'" />
                <div class="mt-4 rounded-r-lg border-l-4 border-green-400 bg-green-50 p-4 dark:bg-green-900/20">
                  <p class="mb-2 font-semibold text-green-700 dark:text-green-300">{{ t('guide.firstRunConfig') }}</p>
                  <ul class="space-y-1 text-green-700 dark:text-green-300">
                    <li>• {{ t('guide.selectDevEnv') }}</li>
                    <li>• {{ t('guide.configCodeGen') }}</li>
                    <li>• {{ t('guide.setGptLevel') }}</li>
                    <li>• {{ t('guide.startAiCoding') }}</li>
                  </ul>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Gemini CLI Section -->
        <div v-show="activeTool === 'gemini-cli'" class="space-y-8">
          <!-- Quick Start -->
          <div class="rounded-2xl bg-gradient-to-r from-green-50 to-emerald-50 p-8 text-center dark:from-green-900/20 dark:to-emerald-900/20">
            <h3 class="mb-4 text-2xl font-bold text-green-900 dark:text-green-100">🚀 Gemini CLI {{ t('guide.quickStart') }}</h3>
            <p class="mb-6 text-green-700 dark:text-green-200">Google AI {{ t('guide.codingAssistant') }}，Gemini 2.5 Pro {{ t('guide.powered') }}</p>
            <div class="flex flex-wrap items-center justify-center gap-8 text-green-600 dark:text-green-300">
              <div class="flex items-center">
                <span class="mr-3 flex h-8 w-8 items-center justify-center rounded-full bg-green-600 font-bold text-white">1</span>
                <span>{{ t('guide.installCli') }}</span>
              </div>
              <div class="flex items-center">
                <span class="mr-3 flex h-8 w-8 items-center justify-center rounded-full bg-green-600 font-bold text-white">2</span>
                <span>{{ t('guide.configureKey') }}</span>
              </div>
              <div class="flex items-center">
                <span class="mr-3 flex h-8 w-8 items-center justify-center rounded-full bg-green-600 font-bold text-white">3</span>
                <span>{{ t('guide.startCoding') }}</span>
              </div>
            </div>
          </div>

          <!-- Platform Tabs -->
          <div class="flex justify-center">
            <div class="inline-flex rounded-xl bg-white p-1 shadow-lg dark:bg-dark-800">
              <button
                v-for="platform in platforms"
                :key="platform.id"
                @click="activePlatform = platform.id"
                :class="[
                  'rounded-lg px-6 py-3 text-sm font-medium transition-all',
                  activePlatform === platform.id
                    ? 'bg-green-600 text-white shadow-md'
                    : 'text-gray-600 hover:bg-gray-100 dark:text-dark-300 dark:hover:bg-dark-700'
                ]"
              >
                {{ platform.icon }} {{ platform.name }}
              </button>
            </div>
          </div>

          <!-- Installation Steps -->
          <div class="rounded-2xl bg-white p-8 shadow-lg dark:bg-dark-800">
            <h3 class="mb-6 flex items-center text-2xl font-bold text-gray-900 dark:text-white">
              <span class="mr-4 flex h-10 w-10 items-center justify-center rounded-full bg-green-600 text-white">
                {{ platformIcons[activePlatform] }}
              </span>
              {{ platformNames[activePlatform] }} {{ t('guide.tutorial') }}
            </h3>

            <!-- Step 1: Install Node.js -->
            <div class="mb-8">
              <h4 class="mb-4 flex items-center text-xl font-semibold text-gray-800 dark:text-gray-200">
                <span class="mr-3 flex h-8 w-8 items-center justify-center rounded-full bg-green-600 text-sm font-bold text-white">1</span>
                {{ t('guide.installNodejs') }}
              </h4>
              <div class="rounded-lg bg-gray-50 p-6 dark:bg-dark-700">
                <p class="mb-4 text-gray-700 dark:text-gray-300">{{ t('guide.sameAsClaudeCode') }}</p>
                <CodeBlock :code="'node --version\nnpm --version'" :title="activePlatform === 'windows' ? 'CMD/PowerShell' : 'Terminal'" />
              </div>
            </div>

            <!-- Step 2: Install Gemini CLI -->
            <div class="mb-8">
              <h4 class="mb-4 flex items-center text-xl font-semibold text-gray-800 dark:text-gray-200">
                <span class="mr-3 flex h-8 w-8 items-center justify-center rounded-full bg-green-600 text-sm font-bold text-white">2</span>
                {{ t('guide.installGemini') }}
              </h4>
              <div class="rounded-lg bg-gray-50 p-6 dark:bg-dark-700">
                <CodeBlock :code="'npm install -g @google/gemini-cli'" :title="activePlatform === 'windows' ? 'CMD/PowerShell' : 'Terminal'" />
              </div>
            </div>

            <!-- Step 3: Configure API -->
            <div class="mb-8">
              <h4 class="mb-4 flex items-center text-xl font-semibold text-gray-800 dark:text-gray-200">
                <span class="mr-3 flex h-8 w-8 items-center justify-center rounded-full bg-green-600 text-sm font-bold text-white">3</span>
                {{ t('guide.configureGemini') }}
              </h4>
              <div class="rounded-lg bg-gray-50 p-6 dark:bg-dark-700">
                <div class="mb-4 rounded-r-lg border-l-4 border-red-400 bg-red-50 p-4 dark:bg-red-900/20">
                  <p class="font-medium text-red-700 dark:text-red-300">
                    {{ t('guide.importantNote') }}{{ t('guide.replaceGeminiToken') }}
                  </p>
                </div>

                <h5 class="mb-3 text-md font-medium text-gray-800 dark:text-gray-200">3.1 {{ t('guide.createGeminiDir') }}</h5>
                <p class="mb-3 text-gray-700 dark:text-gray-300">
                  {{ t('guide.configLocation') }}<code class="rounded bg-gray-200 px-2 py-1 text-sm text-gray-800 dark:bg-dark-600 dark:text-gray-200">{{ activePlatform === 'windows' ? '%USERPROFILE%\\.gemini\\' : '~/.gemini/' }}</code>
                </p>

                <h5 class="mb-3 mt-6 text-md font-medium text-gray-800 dark:text-gray-200">3.2 {{ t('guide.createEnvFile') }}</h5>
                <CodeBlock
                  :code='`GOOGLE_GEMINI_BASE_URL=` + apiBaseUrl + `\nGEMINI_API_KEY=` + t("guide.pasteGeminiToken") + `\nGEMINI_MODEL=gemini-3-pro-preview`'
                  title=".env"
                  :copyText="t('guide.copyConfig')"
                />

                <h5 class="mb-3 mt-6 text-md font-medium text-gray-800 dark:text-gray-200">3.3 {{ t('guide.createSettingsJson') }}</h5>
                <CodeBlock
                  :code='`{\n  "ide": {\n    "enabled": true\n  },\n  "security": {\n    "auth": {\n      "selectedType": "gemini-api-key"\n    }\n  }\n}`'
                  title="settings.json"
                  :copyText="t('guide.copyConfig')"
                />
              </div>
            </div>

            <!-- Step 4: Start Gemini CLI -->
            <div>
              <h4 class="mb-4 flex items-center text-xl font-semibold text-gray-800 dark:text-gray-200">
                <span class="mr-3 flex h-8 w-8 items-center justify-center rounded-full bg-green-600 text-sm font-bold text-white">4</span>
                {{ t('guide.startGemini') }}
              </h4>
              <div class="rounded-lg bg-gray-50 p-6 dark:bg-dark-700">
                <CodeBlock :code="'gemini'" :title="activePlatform === 'windows' ? 'CMD/PowerShell' : 'Terminal'" />
                <div class="mt-4 rounded-r-lg border-l-4 border-green-400 bg-green-50 p-4 dark:bg-green-900/20">
                  <p class="mb-2 font-semibold text-green-700 dark:text-green-300">{{ t('guide.startUsingGemini') }}</p>
                  <ul class="space-y-1 text-green-700 dark:text-green-300">
                    <li>• {{ t('guide.largeContext') }}</li>
                    <li>• {{ t('guide.agentMode') }}</li>
                    <li>• {{ t('guide.googleSearch') }}</li>
                  </ul>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </main>

    <!-- Footer -->
    <footer class="relative z-10 border-t border-gray-200/50 px-6 py-8 dark:border-dark-800/50">
      <div
        class="mx-auto flex max-w-6xl flex-col items-center justify-center gap-4 text-center sm:flex-row sm:text-left"
      >
        <p class="text-sm text-gray-500 dark:text-dark-400">
          &copy; {{ currentYear }} {{ siteName }}. {{ t('home.footer.allRightsReserved') }}
        </p>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import Icon from '@/components/icons/Icon.vue'
import CodeBlock from '@/components/common/CodeBlock.vue'

const { t } = useI18n()

const authStore = useAuthStore()
const appStore = useAppStore()

// Site settings
const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'Sub2API')
const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')
const apiBaseUrl = computed(() => appStore.cachedPublicSettings?.api_base_url || window.location.origin)
const tokenUrl = computed(() => `${window.location.origin}/keys`)

// Theme
const isDark = ref(document.documentElement.classList.contains('dark'))

// Auth state
const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => isAdmin.value ? '/admin/dashboard' : '/dashboard')
const userInitial = computed(() => {
  const user = authStore.user
  if (!user || !user.email) return ''
  return user.email.charAt(0).toUpperCase()
})

// Current year for footer
const currentYear = computed(() => new Date().getFullYear())

// Active tool and platform
const activeTool = ref('claude-code')
const activePlatform = ref('windows')

// Tools configuration
const tools = [
  { id: 'claude-code', name: 'Claude Code', activeClass: 'bg-blue-600' },
  { id: 'codex', name: 'CodeX', activeClass: 'bg-purple-600' },
  { id: 'gemini-cli', name: 'Gemini CLI', activeClass: 'bg-green-600' }
]

// Platforms configuration
const platforms = [
  { id: 'windows', name: 'Windows', icon: '🪟' },
  { id: 'macos', name: 'macOS', icon: '🍎' },
  { id: 'linux', name: 'Linux', icon: '🐧' }
]

const platformIcons: Record<string, string> = {
  windows: '🪟',
  macos: '🍎',
  linux: '🐧'
}

const platformNames: Record<string, string> = {
  windows: 'Windows',
  macos: 'macOS',
  linux: 'Linux'
}

// Toggle theme
function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

// Initialize theme
function initTheme() {
  const savedTheme = localStorage.getItem('theme')
  if (
    savedTheme === 'dark' ||
    (!savedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)
  ) {
    isDark.value = true
    document.documentElement.classList.add('dark')
  }
}

onMounted(() => {
  initTheme()
  authStore.checkAuth()
  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }
})
</script>
