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
        EVENT_CLICK = 'click.' + NAMESPACE,
        EVENT_CHANGE = 'change.' + NAMESPACE,
        EVENT_ENABLE = 'enable.' + NAMESPACE,
        EVENT_DISABLE = 'disable.' + NAMESPACE,
        CLASS_SELECTOR = '.qor-frequency__selector',
        CLASS_MONTH = '.qor-frequency__monthly',
        CLASS_WEEK = '.qor-frequency__weekly',
        CLASS_DAYOFWEEK_INPUT = '.qor-frequency__weekly--input',
        CLASS_DAYOFWEEK = '.qor-frequency__week li';

    function QorMetaFrequency(element, options) {
        this.$element = $(element);
        this.options = $.extend({}, QorMetaFrequency.DEFAULTS, $.isPlainObject(options) && options);
        this.init();
    }

    QorMetaFrequency.prototype = {
        constructor: QorMetaFrequency,

        init: function() {
            let $element = this.$element;
            this.$selector = $element.find(CLASS_SELECTOR);
            this.$month = $element.find(CLASS_MONTH);
            this.$week = $element.find(CLASS_WEEK);
            this.bind();
            this.initData();
        },

        bind: function() {
            this.$element.on(EVENT_CHANGE, CLASS_SELECTOR, this.change.bind(this)).on(EVENT_CLICK, CLASS_DAYOFWEEK, this.selectDayOfWeek.bind(this));
        },

        unbind: function() {
            this.$element.off(EVENT_CHANGE).off(EVENT_CLICK);
        },

        initData: function() {
            this.change();
        },

        change: function() {
            let frequencyVal = this.$selector.val();

            this.$month.hide();
            this.$week.hide();

            if (frequencyVal === 'monthly') {
                this.$month.show();
            } else if (frequencyVal === 'weekly') {
                this.$week.show();
            }
        },

        selectDayOfWeek: function(e) {
            let $target = $(e.target),
                selectWeek = $target.attr('value');

            this.$element.find(CLASS_DAYOFWEEK).removeClass('selected');
            $target.addClass('selected');
            this.$element.find(CLASS_DAYOFWEEK_INPUT).val(selectWeek);
        },

        destroy: function() {
            this.unbind();
            this.$element.removeData(NAMESPACE);
        }
    };

    QorMetaFrequency.plugin = function(options) {
        return this.each(function() {
            var $this = $(this);
            var data = $this.data(NAMESPACE);
            var fn;

            if (!data) {
                if (/destroy/.test(options)) {
                    return;
                }

                $this.data(NAMESPACE, (data = new QorMetaFrequency(this, options)));
            }

            if (typeof options === 'string' && $.isFunction((fn = data[options]))) {
                fn.apply(data);
            }
        });
    };

    $(function() {
        var selector = '[data-toggle="qor.metas.frequency"]';
        $(document)
            .on(EVENT_DISABLE, function(e) {
                QorMetaFrequency.plugin.call($(selector, e.target), 'destroy');
            })
            .on(EVENT_ENABLE, function(e) {
                QorMetaFrequency.plugin.call($(selector, e.target));
            })
            .triggerHandler(EVENT_ENABLE);
    });

    return QorMetaFrequency;
});
