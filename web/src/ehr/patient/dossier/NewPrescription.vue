<template>
  <modal-window
    title="Create new prescription"
    type="add"
    :confirm-fn="submit"
    confirm-text="Create Prescription"
    :cancel-route="{
      name: 'ehr.patient.episode.edit',
      params: { id: $route.params.id, episodeID: $route.params.episodeID },
    }"
  >
    <div class="mt-4">
      <form>
        <form-errors-banner :errors="formErrors" />

        <label>Prescription:</label>
        <input
          type="text"
          disabled
          value="INSULINE INSULATARD INJ 100IE/ML FLACON 10ML"
        />
        <input type="text" v-model="prescription.name" />
      </form>
    </div>
  </modal-window>
</template>

<script>
import ModalWindow from "../../../components/ModalWindow.vue";
import FormErrorsBanner from "../../../components/FormErrorsBanner.vue";

export default {
  name: "NewPrescription",
  components: { ModalWindow, FormErrorsBanner },
  data() {
    return {
      formErrors: [],
      prescription: {
        name: null,
      },
    };
  },
  methods: {
    checkForm(e) {
      // reset the errors
      this.formErrors.length = 0;

      if (!this.prescription.name) {
        this.formErrors.push("Please enter name for prescription");
      }

      return this.formErrors.length === 0;
    },
    submit() {
      if (!this.checkForm()) {
        return false;
      }

      this.loading = true;

      const patientID = this.$route.params.id;

      const payload = {
        type: "prescription",
        patientID,
        value: this.prescription.name.toString(),
        episodeID: this.$route.params.episodeID,
      };

      this.$api.createPrescription({
        body: payload,
        patientID,
      });
      this.$router.push({
        name: "ehr.patient.episode.edit",
        params: { id: this.$route.params.id },
      });
    },
  },
};
</script>
