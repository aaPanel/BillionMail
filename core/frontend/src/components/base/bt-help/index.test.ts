import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import BtHelp from './index.vue'

// Mock naive-ui button
const NButton = {
  template: '<button v-bind="$attrs"><slot /></button>',
  props: ['text'],
}

describe('BtHelp', () => {
  let windowOpenSpy: any

  beforeEach(() => {
    windowOpenSpy = vi.spyOn(window, 'open').mockImplementation(() => null)
  })

  const mountComponent = (props = {}) => {
    return mount(BtHelp, {
      props,
      global: {
        stubs: {
          'n-button': NButton,
        },
        mocks: {
          $t: (key: string) => key,
        },
      },
    })
  }

  it('renders with default props', () => {
    const wrapper = mountComponent()
    expect(wrapper.find('button').exists()).toBe(true)
  })

  it('displays custom text when provided', () => {
    const wrapper = mountComponent({ text: 'Documentation' })
    expect(wrapper.text()).toContain('Documentation')
  })

  it('displays default i18n key text when no text prop', () => {
    const wrapper = mountComponent()
    expect(wrapper.text()).toContain('common.actions.help')
  })

  it('opens href in new window when clicked', async () => {
    const wrapper = mountComponent({ href: 'https://docs.example.com' })
    await wrapper.find('button').trigger('click')
    expect(windowOpenSpy).toHaveBeenCalledWith('https://docs.example.com')
  })

  it('opens empty string when no href provided', async () => {
    const wrapper = mountComponent()
    await wrapper.find('button').trigger('click')
    expect(windowOpenSpy).toHaveBeenCalledWith('')
  })
})
