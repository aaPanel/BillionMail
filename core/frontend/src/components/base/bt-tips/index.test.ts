import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import BtTips from './index.vue'

describe('BtTips', () => {
  it('renders a ul element with bt-tips class', () => {
    const wrapper = mount(BtTips)
    expect(wrapper.find('ul.bt-tips').exists()).toBe(true)
  })

  it('renders slot content', () => {
    const wrapper = mount(BtTips, {
      slots: {
        default: '<li>Tip 1</li><li>Tip 2</li>',
      },
    })
    const items = wrapper.findAll('li')
    expect(items).toHaveLength(2)
    expect(items[0].text()).toBe('Tip 1')
    expect(items[1].text()).toBe('Tip 2')
  })

  it('renders empty when no slot content', () => {
    const wrapper = mount(BtTips)
    expect(wrapper.find('ul').text()).toBe('')
  })
})
