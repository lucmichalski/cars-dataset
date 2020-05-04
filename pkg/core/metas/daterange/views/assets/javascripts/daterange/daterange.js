(function(factory) {
    if (typeof define === 'function' && define.amd) {
        // AMD. Register as anonymous module.
        define(['jquery'], factory);
    } else if (typeof exports === 'object') {
        // Node / CommonJS
        factory(require('jquery'));
    } else {
        // Browser globals.
        factory(jQuery);
    }
})(function($) {
    'use strict';

    let NAMESPACE = 'qor.metas.daterange',
        EVENT_CHANGE = 'change.' + NAMESPACE,
        EVENT_ENABLE = 'enable.' + NAMESPACE,
        EVENT_DISABLE = 'disable.' + NAMESPACE,
        CLASS_SELECTOR = '.qor-daterange__selector',
        CLASS_INPUTS = '.qor-daterange__inputs';

    function QorMetaDateRange(element, options) {
        this.$element = $(element);
        this.options = $.extend({}, QorMetaDateRange.DEFAULTS, $.isPlainObject(options) && options);
        this.init();
    }

    QorMetaDateRange.prototype = {
        constructor: QorMetaDateRange,

        init: function() {
            let $element = this.$element;
            this.$selector = $element.find(CLASS_SELECTOR);
            this.$inputs = $element.find(CLASS_INPUTS);
            this.bind();
            this.initData();
        },

        bind: function() {
            this.$element.on(EVENT_CHANGE, CLASS_SELECTOR, this.change.bind(this));
        },

        unbind: function() {
            this.$element.off(EVENT_CHANGE);
        },

        initData: function() {
            this.change();
        },

        change: function() {
            if (this.$selector.val() === 'custom') {
                this.$inputs.show();
            } else {
                this.$inputs.hide();
            }
        },

        destroy: function() {
            this.unbind();
            this.$element.removeData(NAMESPACE);
        }
    };

    QorMetaDateRange.plugin = function(options) {
        return this.each(function() {
            var $this = $(this);
            var data = $this.data(NAMESPACE);
            var fn;

            if (!data) {
                if (/destroy/.test(options)) {
                    return;
                }

                $this.data(NAMESPACE, (data = new QorMetaDateRange(this, options)));
            }

            if (typeof options === 'string' && $.isFunction((fn = data[options]))) {
                fn.apply(data);
            }
        });
    };

    $(function() {
        var selector = '[data-toggle="qor.metas.daterange"]';
        $(document)
            .on(EVENT_DISABLE, function(e) {
                QorMetaDateRange.plugin.call($(selector, e.target), 'destroy');
            })
            .on(EVENT_ENABLE, function(e) {
                QorMetaDateRange.plugin.call($(selector, e.target));
            })
            .triggerHandler(EVENT_ENABLE);
    });

    return QorMetaDateRange;
});
