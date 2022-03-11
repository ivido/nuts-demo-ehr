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
        <div>
          <label for="medication_select">Choose medication</label>
          <div class="custom-select">
            <select id="medication_select" v-model="selectedMedication">
              <option
                v-for="c in medications"
                v-bind:value="c"
                v-bind:key="c.id"
              >
                {{ c.name }}
              </option>
            </select>

            <svg
              xmlns="http://www.w3.org/2000/svg"
              viewBox="0 0 24 24"
              fill="#444"
            >
              <path d="M24 24H0V0h24v24z" fill="none" opacity=".87" />
              <path
                d="M16.59 8.59L12 13.17 7.41 8.59 6 10l6 6 6-6-1.41-1.41z"
              />
            </svg>
          </div>
        </div>
        <label>Dosage:</label>
        <input type="text" v-model="prescription.dosage.quantity" />
        doses of <input type="text" v-model="prescription.dosage.amount" />
        <select v-model="prescription.dosage.unit">
          <option value="stuk">stuk</option>
          <option value="gr">gr</option>
          <option value="ml">ml</option>
        </select>
        <label>per</label>
        <select v-model="prescription.dosage.frequency">
          <option value="1">day</option>
          <option value="7">week</option>
          <option value="30">month</option>
        </select>
        for <input type="text" v-model="prescription.dosage.period" /> days
        <label>Instructions</label>
        <input type="text" v-model="prescription.dosage.instructions" />
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
      medications: [],
      selectedMedication: null,
      prescription: {
        dosage: {
          quantity: 1,
          amount: 1,
          unit: "stuk",
          frequency: 1,
          period: 7,
          instructions: "",
        },
      },
    };
  },
  created() {
    this.fetchMedication();
  },
  methods: {
    fetchMedication() {
      this.$api
        .getMedications()
        .then((data) => (this.medications = data))
        .catch((response) => {
          console.error("failure", response);
        })
        .finally(() => (this.loading = false));
    },

    checkForm(e) {
      // reset the errors
      this.formErrors.length = 0;

      if (!this.selectedMedication) {
        this.formErrors.push("Please select medication for prescription");
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
        episodeID: this.$route.params.episodeID,
        medicationID: this.selectedMedication.id,
        medicationName: this.selectedMedication.name,
        dosage: {
          quantity: parseInt(this.prescription.dosage.quantity),
          amount: parseInt(this.prescription.dosage.amount),
          unit: this.prescription.dosage.unit,
          frequency: parseInt(this.prescription.dosage.frequency),
          period: parseInt(this.prescription.dosage.period),
          instructions: this.prescription.dosage.instructions,
        },
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
